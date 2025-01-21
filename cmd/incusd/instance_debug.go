package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	internalInstance "github.com/lxc/incus/v6/internal/instance"
	"github.com/lxc/incus/v6/internal/server/instance"
	"github.com/lxc/incus/v6/internal/server/instance/instancetype"
	"github.com/lxc/incus/v6/internal/server/request"
	"github.com/lxc/incus/v6/internal/server/response"
)

// swagger:operation GET /1.0/instances/{name}/debug/memory instances instance_debug_memory_get
//
//	Get memory debug information of an instance
//
//	Returns memory debug information of a running instance.
//	Only supported for VMs.
//
//	---
//	parameters:
//	  - in: path
//	    name: name
//	    description: Instance name
//	    type: string
//	    required: true
//	  - in: query
//	    name: project
//	    description: Project name
//	    type: string
//	    example: default
//	responses:
//	  "200":
//	    description: Success
//	  "400":
//	    $ref: "#/responses/BadRequest"
//	  "403":
//	    $ref: "#/responses/Forbidden"
//	  "404":
//	    $ref: "#/responses/NotFound"
//	  "500":
//	    $ref: "#/responses/InternalServerError"
func instanceDebugMemoryGet(d *Daemon, r *http.Request) response.Response {
	s := d.State()

	path := request.QueryParam(r, "path")
	format := request.QueryParam(r, "format")

	instanceType, err := urlInstanceTypeDetect(r)
	if err != nil {
		return response.SmartError(err)
	}

	projectName := request.ProjectParam(r)
	name, err := url.PathUnescape(mux.Vars(r)["name"])
	if err != nil {
		return response.SmartError(err)
	}

	if internalInstance.IsSnapshot(name) {
		return response.BadRequest(fmt.Errorf("Invalid instance name"))
	}

	// Handle requests targeted to a container on a different node
	resp, err := forwardedResponseIfInstanceIsRemote(s, r, projectName, name, instanceType)
	if err != nil {
		return response.SmartError(err)
	}

	if resp != nil {
		return resp
	}

	// Ensure instance exists.
	inst, err := instance.LoadByProjectAndName(s, projectName, name)
	if err != nil {
		return response.SmartError(err)
	}

	if inst.Type() != instancetype.VM {
		return response.BadRequest(fmt.Errorf("Memory dumps are only supported for virtual machines"))
	}

	if !inst.IsRunning() {
		return response.BadRequest(fmt.Errorf("Instance must be running to dump memory"))
	}

	v, ok := inst.(instance.VM)
	if !ok {
		return response.SmartError(fmt.Errorf("Failed to cast inst to VM"))
	}

	err = v.DumpGuestMemory(path, format)
	if err != nil {
		return response.SmartError(err)
	}

	return response.EmptySyncResponse
}
