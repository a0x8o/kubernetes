<<<<<<< HEAD
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
=======
package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_binary",
    "go_library",
)

go_binary(
    name = "codecgen",
    library = ":go_default_library",
)
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224

go_library(
    name = "go_default_library",
    srcs = [
        "gen.go",
        "z.go",
    ],
<<<<<<< HEAD
    visibility = ["//visibility:private"],
)

go_binary(
    name = "codecgen",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
