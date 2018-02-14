package main

type LockForm struct {
	User      string `form:"user" validate:"required"`
	Operation string `form:"operation" validate:"required,eq=gain|eq=release|eq=extend"`
}
