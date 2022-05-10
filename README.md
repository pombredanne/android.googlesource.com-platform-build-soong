# Soong

Soong is the replacement for the old Android make-based build system.  It
replaces Android.mk files with Android.bp files, which are JSON-like simple
declarative descriptions of modules to build.

See [Simple Build
Configuration](https://source.android.com/compatibility/tests/development/blueprints)
on source.android.com to read how Soong is configured for testing.

## Android.bp file format

By design, Android.bp files are very simple.  There are no conditionals or
control flow statements - any complexity is handled in build logic written in
Go.  The syntax and semantics of Android.bp files are intentionally similar
to [Bazel BUILD files](https://www.bazel.io/versions/master/docs/be/overview.html)
when possible.

### Modules

A module in an Android.bp file starts with a module type, followed by a set of
properties in `name: value,` format:

```
cc_binary {
    name: "gzip",
    srcs: ["src/test/minigzip.c"],
    shared_libs: ["libz"],
    stl: "none",
}
```

Every module must have a `name` property, and the value must be unique across
all Android.bp files.

The list of valid module types and their properties can be generated by calling
`m soong_docs`. It will be written to `$OUT_DIR/soong/docs/soong_build.html`.
This list for the current version of Soong can be found [here](https://ci.android.com/builds/latest/branches/aosp-build-tools/targets/linux/view/soong_build.html).

### File lists

Properties that take a list of files can also take glob patterns and output path
expansions.

* Glob patterns can contain the normal Unix wildcard `*`, for example `"*.java"`.

  Glob patterns can also contain a single `**` wildcard as a path element, which
  will match zero or more path elements. For example, `java/**/*.java` will match
  `java/Main.java` and `java/com/android/Main.java`.

* Output path expansions take the format `:module` or `:module{.tag}`, where
  `module` is the name of a module that produces output files, and it expands to
  a list of those output files. With the optional `{.tag}` suffix, the module
  may produce a different list of outputs according to `tag`.

  For example, a `droiddoc` module with the name "my-docs" would return its
  `.stubs.srcjar` output with `":my-docs"`, and its `.doc.zip` file with
  `":my-docs{.doc.zip}"`.

  This is commonly used to reference `filegroup` modules, whose output files
  consist of their `srcs`.

### Variables

An Android.bp file may contain top-level variable assignments:
```
gzip_srcs = ["src/test/minigzip.c"],

cc_binary {
    name: "gzip",
    srcs: gzip_srcs,
    shared_libs: ["libz"],
    stl: "none",
}
```

Variables are scoped to the remainder of the file they are declared in, as well
as any child Android.bp files.  Variables are immutable with one exception - they
can be appended to with a += assignment, but only before they have been
referenced.

### Comments

Android.bp files can contain C-style multiline `/* */` and C++ style single-line
`//` comments.

### Types

Variables and properties are strongly typed, variables dynamically based on the
first assignment, and properties statically by the module type.  The supported
types are:
* Bool (`true` or `false`)
* Integers (`int`)
* Strings (`"string"`)
* Lists of strings (`["string1", "string2"]`)
* Maps (`{key1: "value1", key2: ["value2"]}`)

Maps may values of any type, including nested maps.  Lists and maps may have
trailing commas after the last value.

Strings can contain double quotes using `\"`, for example `"cat \"a b\""`.

### Operators

Strings, lists of strings, and maps can be appended using the `+` operator.
Integers can be summed up using the `+` operator. Appending a map produces the
union of keys in both maps, appending the values of any keys that are present
in both maps.

### Defaults modules

A defaults module can be used to repeat the same properties in multiple modules.
For example:

```
cc_defaults {
    name: "gzip_defaults",
    shared_libs: ["libz"],
    stl: "none",
}

cc_binary {
    name: "gzip",
    defaults: ["gzip_defaults"],
    srcs: ["src/test/minigzip.c"],
}
```

### Packages

The build is organized into packages where each package is a collection of related files and a
specification of the dependencies among them in the form of modules.

A package is defined as a directory containing a file named `Android.bp`, residing beneath the
top-level directory in the build and its name is its path relative to the top-level directory. A
package includes all files in its directory, plus all subdirectories beneath it, except those which
themselves contain an `Android.bp` file.

The modules in a package's `Android.bp` and included files are part of the module.

For example, in the following directory tree (where `.../android/` is the top-level Android
directory) there are two packages, `my/app`, and the subpackage `my/app/tests`. Note that
`my/app/data` is not a package, but a directory belonging to package `my/app`.

    .../android/my/app/Android.bp
    .../android/my/app/app.cc
    .../android/my/app/data/input.txt
    .../android/my/app/tests/Android.bp
    .../android/my/app/tests/test.cc

This is based on the Bazel package concept.

The `package` module type allows information to be specified about a package. Only a single
`package` module can be specified per package and in the case where there are multiple `.bp` files
in the same package directory it is highly recommended that the `package` module (if required) is
specified in the `Android.bp` file.

Unlike most module type `package` does not have a `name` property. Instead the name is set to the
name of the package, e.g. if the package is in `top/intermediate/package` then the package name is
`//top/intermediate/package`.

E.g. The following will set the default visibility for all the modules defined in the package and
any subpackages that do not set their own default visibility (irrespective of whether they are in
the same `.bp` file as the `package` module) to be visible to all the subpackages by default.

```
package {
    default_visibility: [":__subpackages"]
}
```

### Referencing Modules

A module `libfoo` can be referenced by its name

```
cc_binary {
    name: "app",
    shared_libs: ["libfoo"],
}
```

Obviously, this works only if there is only one `libfoo` module in the source
tree. Ensuring such name uniqueness for larger trees may become problematic. We
might also want to use the same name in multiple mutually exclusive subtrees
(for example, implementing different devices) deliberately in order to describe
a functionally equivalent module. Enter Soong namespaces.

#### Namespaces

The presence of the `soong_namespace {..}` in an Android.bp file defines a
**namespace**. For instance, having

```
soong_namespace {
    ...
}
...
```

in `device/google/bonito/Android.bp` informs Soong that within the
`device/google/bonito` package the module names are unique, that is, all the
modules defined in the Android.bp files in the `device/google/bonito/` tree have
unique names. However, there may be modules with the same names outside
`device/google/bonito` tree. Indeed, there is a module `"pixelstats-vendor"`
both in `device/google/bonito/pixelstats` and in
`device/google/coral/pixelstats`.

The name of a namespace is the path of its directory. The name of the namespace
in the example above is thus `device/google/bonito`.

An implicit **global namespace** corresponds to the source tree as a whole. It
has empty name.

A module name's **scope** is the smallest namespace containing it. Suppose a
source tree has `device/my` and `device/my/display` namespaces. If `libfoo`
module is defined in `device/my/display/lib/Android.bp`, its namespace is
`device/my/display`.

The name uniqueness thus means that module's name is unique within its scope. In
other words, "//_scope_:_name_" is globally unique module reference, e.g,
`"//device/google/bonito:pixelstats-vendor"`. _Note_ that the name of the
namespace for a module may be different from module's package name: `libfoo`
belongs to `device/my/display` namespace but is contained in
`device/my/display/lib` package.

#### Name Resolution

The form of a module reference determines how Soong locates the module.

For a **global reference** of the "//_scope_:_name_" form, Soong verifies there
is a namespace called "_scope_", then verifies it contains a "_name_" module and
uses it. Soong verifies there is only one "_name_" in "_scope_" at the beginning
when it parses Android.bp files.

A **local reference** has "_name_" form, and resolving it involves looking for a
module "_name_" in one or more namespaces. By default only the global namespace
is searched for "_name_" (in other words, only the modules not belonging to an
explicitly defined scope are considered). The `imports` attribute of the
`soong_namespaces` allows to specify where to look for modules . For instance,
with `device/google/bonito/Android.bp` containing

```
soong_namespace {
    imports: [
        "hardware/google/interfaces",
        "hardware/google/pixel",
        "hardware/qcom/bootctrl",
    ],
}
```

a reference to `"libpixelstats"` will resolve to the module defined in
`hardware/google/pixel/pixelstats/Android.bp` because this module is in
`hardware/google/pixel` namespace.

**TODO**: Conventionally, languages with similar concepts provide separate
constructs for namespace definition and name resolution (`namespace` and `using`
in C++, for instance). Should Soong do that, too?

#### Referencing modules in makefiles

While we are gradually converting makefiles to Android.bp files, Android build
is described by a mixture of Android.bp and Android.mk files, and a module
defined in an Android.mk file can reference a module defined in Android.bp file.
For instance, a binary still defined in an Android.mk file may have a library
defined in already converted Android.bp as a dependency.

A module defined in an Android.bp file and belonging to the global namespace can
be referenced from a makefile without additional effort. If a module belongs to
an explicit namespace, it can be referenced from a makefile only after after the
name of the namespace has been added to the value of PRODUCT_SOONG_NAMESPACES
variable.

Note that makefiles have no notion of namespaces and exposing namespaces with
the same modules via PRODUCT_SOONG_NAMESPACES may cause Make failure. For
instance, exposing both `device/google/bonito` and `device/google/coral`
namespaces will cause Make failure because it will see two targets for the
`pixelstats-vendor` module.

### Visibility

The `visibility` property on a module controls whether the module can be
used by other packages. Modules are always visible to other modules declared
in the same package. This is based on the Bazel visibility mechanism.

If specified the `visibility` property must contain at least one rule.

Each rule in the property must be in one of the following forms:
* `["//visibility:public"]`: Anyone can use this module.
* `["//visibility:private"]`: Only rules in the module's package (not its
subpackages) can use this module.
* `["//visibility:override"]`: Discards any rules inherited from defaults or a
creating module. Can only be used at the beginning of a list of visibility
rules.
* `["//some/package:__pkg__", "//other/package:__pkg__"]`: Only modules in
`some/package` and `other/package` (defined in `some/package/*.bp` and
`other/package/*.bp`) have access to this module. Note that sub-packages do not
have access to the rule; for example, `//some/package/foo:bar` or
`//other/package/testing:bla` wouldn't have access. `__pkg__` is a special
module and must be used verbatim. It represents all of the modules in the
package.
* `["//project:__subpackages__", "//other:__subpackages__"]`: Only modules in
packages `project` or `other` or in one of their sub-packages have access to
this module. For example, `//project:rule`, `//project/library:lib` or
`//other/testing/internal:munge` are allowed to depend on this rule (but not
`//independent:evil`)
* `["//project"]`: This is shorthand for `["//project:__pkg__"]`
* `[":__subpackages__"]`: This is shorthand for `["//project:__subpackages__"]`
where `//project` is the module's package, e.g. using `[":__subpackages__"]` in
`packages/apps/Settings/Android.bp` is equivalent to
`//packages/apps/Settings:__subpackages__`.
* `["//visibility:legacy_public"]`: The default visibility, behaves as
`//visibility:public` for now. It is an error if it is used in a module.

The visibility rules of `//visibility:public` and `//visibility:private` cannot
be combined with any other visibility specifications, except
`//visibility:public` is allowed to override visibility specifications imported
through the `defaults` property.

Packages outside `vendor/` cannot make themselves visible to specific packages
in `vendor/`, e.g. a module in `libcore` cannot declare that it is visible to
say `vendor/google`, instead it must make itself visible to all packages within
`vendor/` using `//vendor:__subpackages__`.

If a module does not specify the `visibility` property then it uses the
`default_visibility` property of the `package` module in the module's package.

If the `default_visibility` property is not set for the module's package then
it will use the `default_visibility` of its closest ancestor package for which
a `default_visibility` property is specified.

If no `default_visibility` property can be found then the module uses the
global default of `//visibility:legacy_public`.

The `visibility` property has no effect on a defaults module although it does
apply to any non-defaults module that uses it. To set the visibility of a
defaults module, use the `defaults_visibility` property on the defaults module;
not to be confused with the `default_visibility` property on the package module.

Once the build has been completely switched over to soong it is possible that a
global refactoring will be done to change this to `//visibility:private` at
which point all packages that do not currently specify a `default_visibility`
property will be updated to have
`default_visibility = [//visibility:legacy_public]` added. It will then be the
owner's responsibility to replace that with a more appropriate visibility.

### Formatter

Soong includes a canonical formatter for Android.bp files, similar to
[gofmt](https://golang.org/cmd/gofmt/).  To recursively reformat all Android.bp files
in the current directory:
```
bpfmt -w .
```

The canonical format includes 4 space indents, newlines after every element of a
multi-element list, and always includes a trailing comma in lists and maps.

### Convert Android.mk files

Soong includes a tool perform a first pass at converting Android.mk files
to Android.bp files:

```
androidmk Android.mk > Android.bp
```

The tool converts variables, modules, comments, and some conditionals, but any
custom Makefile rules, complex conditionals or extra includes must be converted
by hand.

#### Differences between Android.mk and Android.bp

* Android.mk files often have multiple modules with the same name (for example
for static and shared version of a library, or for host and device versions).
Android.bp files require unique names for every module, but a single module can
be built in multiple variants, for example by adding `host_supported: true`.
The androidmk converter will produce multiple conflicting modules, which must
be resolved by hand to a single module with any differences inside
`target: { android: { }, host: { } }` blocks.

### Conditionals

Soong deliberately does not support most conditionals in Android.bp files.  We
suggest removing most conditionals from the build.  See
[Best Practices](docs/best_practices.md#removing-conditionals) for some
examples on how to remove conditionals.

Most conditionals supported natively by Soong are converted to a map
property.  When building the module one of the properties in the map will be
selected, and its values appended to the property with the same name at the
top level of the module.

For example, to support architecture specific files:
```
cc_library {
    ...
    srcs: ["generic.cpp"],
    arch: {
        arm: {
            srcs: ["arm.cpp"],
        },
        x86: {
            srcs: ["x86.cpp"],
        },
    },
}
```

When building the module for arm the `generic.cpp` and `arm.cpp` sources will
be built.  When building for x86 the `generic.cpp` and 'x86.cpp' sources will
be built.

#### Soong Config Variables

When converting vendor modules that contain conditionals, simple conditionals
can be supported through Soong config variables using `soong_config_*`
modules that describe the module types, variables and possible values:

```
soong_config_module_type {
    name: "acme_cc_defaults",
    module_type: "cc_defaults",
    config_namespace: "acme",
    variables: ["board"],
    bool_variables: ["feature"],
    value_variables: ["width"],
    properties: ["cflags", "srcs"],
}

soong_config_string_variable {
    name: "board",
    values: ["soc_a", "soc_b", "soc_c"],
}
```

This example describes a new `acme_cc_defaults` module type that extends the
`cc_defaults` module type, with three additional conditionals based on
variables `board`, `feature` and `width`, which can affect properties `cflags`
and `srcs`. Additionally, each conditional will contain a `conditions_default`
property can affect `cflags` and `srcs` in the following conditions:

* bool variable (e.g. `feature`): the variable is unspecified or not set to a true value
* value variable (e.g. `width`): the variable is unspecified
* string variable (e.g. `board`): the variable is unspecified or the variable is set to a string unused in the
given module. For example, with `board`, if the `board`
conditional contains the properties `soc_a` and `conditions_default`, when
board=soc_b, the `cflags` and `srcs` values under `conditions_default` will be
used. To specify that no properties should be amended for `soc_b`, you can set
`soc_b: {},`.

The values of the variables can be set from a product's `BoardConfig.mk` file:
```
$(call soong_config_set,acme,board,soc_a)
$(call soong_config_set,acme,feature,true)
$(call soong_config_set,acme,width,200)
```

The `acme_cc_defaults` module type can be used anywhere after the definition in
the file where it is defined, or can be imported into another file with:
```
soong_config_module_type_import {
    from: "device/acme/Android.bp",
    module_types: ["acme_cc_defaults"],
}
```

It can used like any other module type:
```
acme_cc_defaults {
    name: "acme_defaults",
    cflags: ["-DGENERIC"],
    soong_config_variables: {
        board: {
            soc_a: {
                cflags: ["-DSOC_A"],
            },
            soc_b: {
                cflags: ["-DSOC_B"],
            },
            conditions_default: {
                cflags: ["-DSOC_DEFAULT"],
            },
        },
        feature: {
            cflags: ["-DFEATURE"],
            conditions_default: {
                cflags: ["-DFEATURE_DEFAULT"],
            },
        },
        width: {
            cflags: ["-DWIDTH=%s"],
            conditions_default: {
                cflags: ["-DWIDTH=DEFAULT"],
            },
        },
    },
}

cc_library {
    name: "libacme_foo",
    defaults: ["acme_defaults"],
    srcs: ["*.cpp"],
}
```

With the `BoardConfig.mk` snippet above, `libacme_foo` would build with
`cflags: "-DGENERIC -DSOC_A -DFEATURE -DWIDTH=200"`.

Alternatively, with `DefaultBoardConfig.mk`:

```
SOONG_CONFIG_NAMESPACES += acme
SOONG_CONFIG_acme += \
    board \
    feature \
    width \

SOONG_CONFIG_acme_feature := false
```

then `libacme_foo` would build with `cflags: "-DGENERIC -DSOC_DEFAULT -DFEATURE_DEFAULT -DSIZE=DEFAULT"`.

Alternatively, with `DefaultBoardConfig.mk`:

```
SOONG_CONFIG_NAMESPACES += acme
SOONG_CONFIG_acme += \
    board \
    feature \
    width \

SOONG_CONFIG_acme_board := soc_c
```

then `libacme_foo` would build with `cflags: "-DGENERIC -DSOC_DEFAULT
-DFEATURE_DEFAULT -DSIZE=DEFAULT"`.

`soong_config_module_type` modules will work best when used to wrap defaults
modules (`cc_defaults`, `java_defaults`, etc.), which can then be referenced
by all of the vendor's other modules using the normal namespace and visibility
rules.

## Build logic

The build logic is written in Go using the
[blueprint](http://godoc.org/github.com/google/blueprint) framework.  Build
logic receives module definitions parsed into Go structures using reflection
and produces build rules.  The build rules are collected by blueprint and
written to a [ninja](http://ninja-build.org) build file.

## Environment Variables Config File

Soong can optionally load environment variables from a pre-specified
configuration file during startup. These environment variables can be used
to control the behavior of the build. For example, these variables can determine
whether remote-execution should be used for the build or not.

The `ANDROID_BUILD_ENVIRONMENT_CONFIG_DIR` environment variable specifies the
directory in which the config file should be searched for. The
`ANDROID_BUILD_ENVIRONMENT_CONFIG` variable determines the name of the config
file to be searched for within the config directory. For example, the following
build comand will load `ENV_VAR_1` and `ENV_VAR_2` environment variables from
the `example_config.json` file inside the `build/soong` directory.

```
ANDROID_BUILD_ENVIRONMENT_CONFIG_DIR=build/soong \
  ANDROID_BUILD_ENVIRONMENT_CONFIG=example_config \
  build/soong/soong_ui.bash
```

## Other documentation

* [Best Practices](docs/best_practices.md)
* [Build Performance](docs/perf.md)
* [Generating CLion Projects](docs/clion.md)
* [Generating YouCompleteMe/VSCode compile\_commands.json file](docs/compdb.md)
* Make-specific documentation: [build/make/README.md](https://android.googlesource.com/platform/build/+/master/README.md)

## Developing for Soong

To load the code of Soong in IntelliJ:

* File -> Open, open the `build/soong` directory. It will be opened as a new
  project.
* File -> Settings, then Languages & Frameworks -> Go -> GOROOT, then set it to
  `prebuilts/go/linux-x86`
* File -> Project Structure, then, Project Settings -> Modules, then Add
  Content Root, then add the `build/blueprint` directory.
* Optional: also add the `external/golang-protobuf` directory. In practice,
  IntelliJ seems to work well enough without this, too.
### Running Soong in a debugger

To make `soong_build` wait for a debugger connection, install `dlv` and then
start the build with `SOONG_DELVE=<listen addr>` in the environment.
For example:
```bash
SOONG_DELVE=:5006 m nothing
```

To make `soong_ui` wait for a debugger connection, use the `SOONG_UI_DELVE`
variable:

```
SOONG_UI_DELVE=:5006 m nothing
```


setting or unsetting `SOONG_DELVE` causes a recompilation of `soong_build`. This
is because in order to debug the binary, it needs to be built with debug
symbols.

To test the debugger connection, run this command:

```
dlv connect :5006
```

If you see an error:
```
Could not attach to pid 593: this could be caused by a kernel
security setting, try writing "0" to /proc/sys/kernel/yama/ptrace_scope
```
you can temporarily disable
[Yama's ptrace protection](https://www.kernel.org/doc/Documentation/security/Yama.txt)
using:
```bash
sudo sysctl -w kernel.yama.ptrace_scope=0
```

To connect to the process using IntelliJ:

* Run -> Edit Configurations...
* Choose "Go Remote" on the left
* Click on the "+" buttion on the top-left
* Give it a nice name and set "Host" to localhost and "Port" to the port in the
  environment variable

Debugging works far worse than debugging Java, but is sometimes useful.

Sometimes the `dlv` process hangs on connection. A symptom of this is `dlv`
spinning a core or two. In that case, `kill -9` `dlv` and try again.
Anecdotally, it _feels_ like waiting a minute after the start of `soong_build`
helps.

## Contact

Email android-building@googlegroups.com (external) for any questions, or see
[go/soong](http://go/soong) (internal).
