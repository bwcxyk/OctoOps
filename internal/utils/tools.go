/*
@Author : YaoKun
@Time : 2025/7/10 上午9:34
*/

package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// 从中间件上下文获取当前用户的 uid, roleId, 类型是 interface

func GetContextUser(c *gin.Context) (int, int, error) {

	// 先获取接口数据, 再通过断言解析
	uidInterface, ok := c.Get("uid")
	if !ok {
		return 0, 0, errors.New("用户认证失败")
	}
	uid, ok := uidInterface.(int)
	if !ok {
		return 0, 0, errors.New("用户认证失败")
	}

	// roleId
	roleIdInterface, ok := c.Get("roleId")
	if !ok {
		return 0, 0, errors.New("用户权限不足")
	}

	roleId, ok := roleIdInterface.(int)
	if !ok {
		return 0, 0, errors.New("用户权限解析失败")
	}

	return uid, roleId, nil
}
