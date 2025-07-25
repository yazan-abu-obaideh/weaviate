//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2025 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

// Code generated by go-swagger; DO NOT EDIT.

package schema

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/weaviate/weaviate/entities/models"
)

// AliasesDeleteHandlerFunc turns a function with the right signature into a aliases delete handler
type AliasesDeleteHandlerFunc func(AliasesDeleteParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn AliasesDeleteHandlerFunc) Handle(params AliasesDeleteParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// AliasesDeleteHandler interface for that can handle valid aliases delete params
type AliasesDeleteHandler interface {
	Handle(AliasesDeleteParams, *models.Principal) middleware.Responder
}

// NewAliasesDelete creates a new http.Handler for the aliases delete operation
func NewAliasesDelete(ctx *middleware.Context, handler AliasesDeleteHandler) *AliasesDelete {
	return &AliasesDelete{Context: ctx, Handler: handler}
}

/*
	AliasesDelete swagger:route DELETE /aliases/{aliasName} schema aliasesDelete

# Delete an alias

Remove an existing alias from the system. This will delete the alias mapping but will not affect the underlying collection (class).
*/
type AliasesDelete struct {
	Context *middleware.Context
	Handler AliasesDeleteHandler
}

func (o *AliasesDelete) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewAliasesDeleteParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal *models.Principal
	if uprinc != nil {
		principal = uprinc.(*models.Principal) // this is really a models.Principal, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
