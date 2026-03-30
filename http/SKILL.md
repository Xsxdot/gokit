这份基于 `resty` 二次封装的 HTTP 客户端代码，经过优化后，不仅解决了多处潜在隐患，API 的语义也变得极其清晰。

这里为你整理了一份面向开发者的**使用指南（Skills / Cheat Sheet）**。你可以直接将这份文档作为该 package 的 `README.md` 或内部 Wiki，方便团队成员（或你自己）快速上手。

---

# HTTP Client 封装库使用指南 (Skills)

本库提供了一套高度连贯的、支持链式调用的 HTTP 客户端。它完美融合了极简的请求构造、灵活的断言机制以及强大的全局拦截器。

## 💡 1. 基础请求与数据绑定 (Basic & Bind)

最常见的场景：发起请求，并直接将 JSON 响应绑定到结构体。如果发生网络错误或解析错误，会直接通过 `err` 抛出。

### 发起 GET 请求（使用 Struct 构建 Query 参数）
```go
type QueryParams struct {
    Page int    `json:"page"`
    Size int    `json:"size"`
    Tag  string `json:"tag"`
}

var result MyResultStruct

// QueryParamsStruct 会安全地将 Struct 转为 URL 参数，且不会丢失大整数精度
err := http.Get("https://api.example.com/data").
    QueryParamsStruct(QueryParams{Page: 1, Size: 20, Tag: "golang"}).
    Do().
    Bind(&result)

if err != nil {
    log.Fatalf("请求失败: %v", err)
}
```

### 发起 POST JSON 请求
```go
reqData := map[string]interface{}{"name": "test", "role": "admin"}
var result MyResultStruct

// PostJSON 会自动设置 Content-Type: application/json 并序列化 Body
err := http.PostJSON("https://api.example.com/users", reqData).
    Do().
    Bind(&result)
```

---

## 🔍 2. 链式断言 (Chainable Assertions)

非常适合编写**API 自动化测试**，或调用**非标准结构的第三方接口**。通过 `.Ensure...` 建立期望，最后用 `.Unwrap()` 或 `.Err()` 终结链条并获取结果。

```go
// 请求发起 -> 拿到响应 -> 链式断言 -> 获取原始 Body 字节
bodyBytes, err := http.Get("https://api.github.com/users/octocat").
    Do().
    EnsureStatus2xx().                    // 确保状态码是 200-299
    EnsureJsonStringEq("login", "octocat").// 确保 JSON 中的 login 字段等于期望值
    EnsureJsonExists("id").               // 确保存在 id 字段
    Unwrap()                              // 终结断言，提取 []byte

if err != nil {
    // 这里的 err 会自动包含精简的上下文（URL、状态码、截断的 Body），防止日志刷屏
    log.Fatalf("接口校验失败: %v", err)
}
```

---

## ⚡️ 3. 局部数据提取 (Gson 快捷查询)

当你面对一个非常庞大且嵌套极深的 JSON 返回值，但你**只需要其中的一两个字段**时，无需定义庞大的结构体，直接使用内置的 `.Gson()`。

*注：`Gson()` 底层具备缓存机制，多次链式调用或取值不会重复解析 JSON，性能极佳。*

```go
resp := http.Get("https://api.example.com/complex-json").Do()
if err := resp.Err(); err != nil {
    return err
}

// 直接使用 GJSON 语法提取深层嵌套的数据
userID := resp.Gson().Get("data.user.id").Int()
userName := resp.Gson().Get("data.user.profile.name").String()

fmt.Printf("User: %d - %s\n", userID, userName)
```

---

## 🛠️ 4. 统一拦截器：对抗厂商的“统一外壳” (Global Hook)

针对那些总是返回 `{"code": 0, "msg": "success", "data": {...}}` 的厂商 API，**不要在业务代码里反复写断言**。使用独立的 Client 配合 `OnAfterResponse` 拦截器，直接在底层剥离外壳。

### 第一步：初始化专属客户端
```go
// 全局单例，专用于请求某厂商 API
var VendorClient = http.NewClient(nil).OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
    if r.IsError() {
        return nil // 网络层面/非 2xx 的错误，直接放行给上层
    }
    
    // 统一剥壳与业务级错误拦截
    body := r.Body()
    code := gjson.GetBytes(body, "code").Int()
    if code != 0 {
        msg := gjson.GetBytes(body, "msg").String()
        return fmt.Errorf("厂商API业务异常: code=%d, msg=%s", code, msg)
    }
    return nil // 校验通过
})
```

### 第二步：在业务中享受极简调用
```go
var actualData UserData
// 请求发起后，会自动经过 Hook 校验 code == 0。
// 失败会直接返回精简的 error；成功则直接解析 data。
err := VendorClient.Get("/api/v1/user/123").Do().Bind(&actualData)
if err != nil {
    log.Printf("获取失败: %v", err)
}
```

---

## ⚙️ 5. 高级网络配置 (Options)

对于需要配置超时、代理、自定义证书的复杂场景，可以通过 `Options` 构造专用的 Client。配置支持被 Client 下的所有 Request 自动 `Clone` 继承，完全**线程安全**。

```go
// 1. 组装复杂的配置
opts := http.NewOptions().
    WithTimeout(15 * time.Second).
    WithHeader("Authorization", "Bearer token...").
    WithProxy("http://127.0.0.1:7890").
    WithInsecureSkipVerify(true) // 忽略自签证书报错

// 2. 生成客户端
myClient := http.NewClient(opts)

// 3. 并发安全地发起请求 (内部会自动 Clone options)
go myClient.Get("https://xxx.com/api/1").Do()
go myClient.Get("https://xxx.com/api/2").Header("X-Extra", "1").Do()
```

---

### 💡 核心 API 快速查阅速记
* **请求构建**：`.Header()`, `.Cookies()`, `.QueryParams()`, `.QueryParamsStruct()`, `.JSON()`
* **响应终结**：`.Err()` (仅检查错误), `.Unwrap()` (取 byte), `.Bind(&v)` (解析 JSON), `.String()` (取字符串)
* **响应取值**：`.StatusCode()`, `.Headers()` (含多值), `.HeadersFlat()` (单值), `.Gson()`