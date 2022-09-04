// Author: Daniel TAN
// Date: 2021-08-18 23:32:00
// LastEditors: Daniel TAN
// LastEditTime: 2021-12-17 23:50:47
// FilePath: /trinity-micro/core/ioc/container/consts.go
// Description:
package container

import (
	"context"
	"fmt"
)

type Keyword string

const (
	_CONTAINER      Keyword = "container"
	_AUTO_WIRE      Keyword = "autowire"
	_RESOURCE       Keyword = "resource"
	TAG_SPLITTER            = ";"
	TAG_KV_SPLITTER         = ":"
	CONTEXT                 = "CONTEXT"
)

type InstanceName string

func (i InstanceName) Validate(ctx context.Context) error {
	if i == "" {
		return fmt.Errorf("instance name cannot be empty")
	}
	return nil
}

type InstanceType string

const (
	Singleton     InstanceType = "SINGLETON"
	MultiInstance InstanceType = "MULTI_INSTANCE"
)
