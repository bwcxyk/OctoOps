package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"octoops/internal/config"
	"octoops/internal/db"
	"time"
)

const baseURL = "http://localhost:8080/api"

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token       string   `json:"token"`
		User        UserInfo `json:"user"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
	} `json:"data"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Status   int    `json:"status"`
}

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	// 初始化配置和数据库
	config.InitConfig()
	db.Init()

	fmt.Println("开始测试RBAC系统...")

	// 测试登录
	token := testLogin()
	if token == "" {
		fmt.Println("登录失败，退出测试")
		return
	}

	// 测试获取用户信息
	testGetProfile(token)

	// 测试获取用户权限
	testGetPermissions(token)

	// 测试用户管理API
	testUserManagement(token)

	// 测试角色管理API
	testRoleManagement(token)

	// 测试权限管理API
	testPermissionManagement(token)

	fmt.Println("RBAC系统测试完成！")
}

func testLogin() string {
	fmt.Println("\n=== 测试用户登录 ===")

	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	if loginResp.Code == 200 {
		fmt.Printf("登录成功！用户: %s, 角色: %v\n", loginResp.Data.User.Username, loginResp.Data.Roles)
		return loginResp.Data.Token
	} else {
		fmt.Printf("登录失败: %s\n", loginResp.Message)
		return ""
	}
}

func testGetProfile(token string) {
	fmt.Println("\n=== 测试获取用户信息 ===")

	req, _ := http.NewRequest("GET", baseURL+"/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp ApiResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Code == 200 {
		fmt.Println("获取用户信息成功")
	} else {
		fmt.Printf("获取用户信息失败: %s\n", apiResp.Message)
	}
}

func testGetPermissions(token string) {
	fmt.Println("\n=== 测试获取用户权限 ===")

	req, _ := http.NewRequest("GET", baseURL+"/auth/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("获取用户权限失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp ApiResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Code == 200 {
		fmt.Println("获取用户权限成功")
	} else {
		fmt.Printf("获取用户权限失败: %s\n", apiResp.Message)
	}
}

func testUserManagement(token string) {
	fmt.Println("\n=== 测试用户管理API ===")

	// 测试获取用户列表
	req, _ := http.NewRequest("GET", baseURL+"/users?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("获取用户列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp ApiResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Code == 200 {
		fmt.Println("获取用户列表成功")
	} else {
		fmt.Printf("获取用户列表失败: %s\n", apiResp.Message)
	}
}

func testRoleManagement(token string) {
	fmt.Println("\n=== 测试角色管理API ===")

	// 测试获取角色列表
	req, _ := http.NewRequest("GET", baseURL+"/roles?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("获取角色列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp ApiResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Code == 200 {
		fmt.Println("获取角色列表成功")
	} else {
		fmt.Printf("获取角色列表失败: %s\n", apiResp.Message)
	}
}

func testPermissionManagement(token string) {
	fmt.Println("\n=== 测试权限管理API ===")

	// 测试获取权限列表
	req, _ := http.NewRequest("GET", baseURL+"/permissions?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("获取权限列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp ApiResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Code == 200 {
		fmt.Println("获取权限列表成功")
	} else {
		fmt.Printf("获取权限列表失败: %s\n", apiResp.Message)
	}

	// 测试获取权限树
	req, _ = http.NewRequest("GET", baseURL+"/permissions/tree", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("获取权限树失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if apiResp.Code == 200 {
		fmt.Println("获取权限树成功")
	} else {
		fmt.Printf("获取权限树失败: %s\n", apiResp.Message)
	}
}

// 辅助函数：等待服务启动
func waitForService() {
	fmt.Println("等待服务启动...")
	for i := 0; i < 30; i++ {
		resp, err := http.Get("http://localhost:8080/api/auth/login")
		if err == nil {
			resp.Body.Close()
			fmt.Println("服务已启动")
			return
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("服务启动超时")
}
