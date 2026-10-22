//go:generate swag init -d ../../../cmd/server,./ --pd -o ./ -ot go

package handler

import httpSwagger "github.com/swaggo/http-swagger"

var SwaggerHandler = httpSwagger.WrapHandler
