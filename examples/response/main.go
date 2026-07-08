package main

import (
	foundationErrors "github.com/jackman0925/go-foundation/errors"
	"github.com/jackman0925/go-foundation/response"
)

func main() {
	err := foundationErrors.New(foundationErrors.CodeBadRequest, "参数错误")
	resp := response.NewFail(err)
	_ = resp
}
