load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "org_golang_x_exp",
        importpath = "golang.org/x/exp",
        sum = "h1:TfdoLivD44QwvssI9Sv1xwa5DcL5XQr4au4sZ2F2NV4=",
        version = "v0.0.0-20220428152302-39d4317da171",
    )
    go_repository(
        name = "org_golang_x_mod",
        importpath = "golang.org/x/mod",
        sum = "h1:LQmS1nU0twXLA96Kt7U9qtHJEbBk3z6Q0V4UXjZkpr4=",
        version = "v0.6.0-dev.0.20211013180041-c96bc1413d57",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:id054HUawV2/6IGm2IV8KZQjqtwAOo2CYlOToYqa0d0=",
        version = "v0.0.0-20211019181941-9d821ace8654",
    )
    go_repository(
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:0c3L82FDQ5rt1bjTBlchS8t6RQ6299/+5bWMnRLh+uI=",
        version = "v0.1.8-0.20211029000441-d6a9af8af023",
    )
    go_repository(
        name = "org_golang_x_xerrors",
        importpath = "golang.org/x/xerrors",
        sum = "h1:go1bK/D/BFZV2I8cIQd1NKEZ+0owSTG1fDTci4IqFcE=",
        version = "v0.0.0-20200804184101-5ec99f83aff1",
    )
