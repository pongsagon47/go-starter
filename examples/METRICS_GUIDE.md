# 📊 คู่มือการใช้งาน Metrics แบบเข้าใจง่าย

## 🤔 Metrics คืออะไร?

**Metrics** คือการวัดและเก็บข้อมูลต่าง ๆ ของระบบ เพื่อให้เราทราบว่าแอปของเราทำงานเป็นอย่างไร

เปรียบเทียบง่าย ๆ:

- **เหมือนมาตรวัดในรถยนต์** - ความเร็ว, อุณหภูมิเครื่องยนต์, น้ำมัน
- **เหมือนการตรวจสุขภาพ** - อุณหภูมิ, ความดัน, น้ำหนัก

## 🎯 ทำไมต้องใช้ Metrics?

1. **รู้ว่าระบบทำงานดีไหม** - มี error หรือไม่? ช้าเร็วแค่ไหน?
2. **วางแผนการขยาย** - จำนวน user เพิ่มขึ้น ต้องเพิ่ม server หรือไม่?
3. **แก้ปัญหาได้เร็ว** - รู้ทันทีว่าจุดไหนเกิดปัญหา
4. **วิเคราะห์ธุรกิจ** - มียอดขายเท่าไหร่? user มากน้อยแค่ไหน?

## 📏 ประเภทของ Metrics (4 แบบหลัก)

### 1. **Counter** - ตัวนับที่เพิ่มเรื่อย ๆ

```go
// ตัวอย่าง: นับจำนวนคำขอ HTTP
requestCounter := metrics.NewCounter("http_requests_total", "จำนวนคำขอทั้งหมด", labels)
requestCounter.Inc() // เพิ่ม 1
requestCounter.Add(5) // เพิ่ม 5
```

**ใช้กับ:** จำนวน request, จำนวน error, จำนวน user ที่สมัคร

### 2. **Gauge** - ค่าที่ขึ้นลงได้

```go
// ตัวอย่าง: จำนวน user ออนไลน์
activeUsers := metrics.NewGauge("active_users", "ผู้ใช้ออนไลน์", labels)
activeUsers.Set(100) // ตั้งค่าเป็น 100
activeUsers.Inc()    // เพิ่ม 1 (เป็น 101)
activeUsers.Dec()    // ลด 1 (เป็น 100)
```

**ใช้กับ:** จำนวน connection, CPU usage, Memory usage, จำนวน user ออนไลน์

### 3. **Histogram** - วัดการกระจายของค่า

```go
// ตัวอย่าง: เวลาในการตอบสนอง
responseTime := metrics.NewHistogram("response_time_seconds", "เวลาตอบสนอง", nil, labels)
responseTime.Observe(0.25) // บันทึกว่าใช้เวลา 0.25 วินาที
```

**ใช้กับ:** เวลาในการตอบสนอง, ขนาดของไฟล์, ราคาสินค้า

### 4. **Timer** - วัดเวลาแบบสะดวก

```go
// ตัวอย่าง: วัดเวลาการประมวลผล
timer := metrics.NewTimer(responseTime)
stop := timer.Start()
// ... ทำงานต่าง ๆ ...
stop() // บันทึกเวลาอัตโนมัติ
```

## 🏷️ Labels - การติดป้ายกำกับ

Labels ช่วยแยกแยะข้อมูลได้ละเอียดขึ้น:

```go
// ไม่มี label - ข้อมูลรวม ๆ
requestCounter.Inc()

// มี label - แยกตาม method และ status
requestCounter.With(metrics.Labels{
    "method": "GET",
    "status": "200",
    "endpoint": "/api/users",
}).Inc()
```

**ผลลัพธ์:**

- `http_requests_total{method="GET", status="200", endpoint="/api/users"}` = 150
- `http_requests_total{method="POST", status="500", endpoint="/api/users"}` = 5

## 🔧 การใช้งานในแอป (ตัวอย่างง่าย ๆ)

### 1. ตั้งค่าเบื้องต้น

```go
// ตั้งค่า metrics สำหรับแอป
config := &metrics.Config{
    Enabled:   true,
    Namespace: "myshop",  // ชื่อแอป
    Subsystem: "api",     // ส่วนย่อย
}
metrics.SetDefaultConfig(config)
```

### 2. เพิ่ม Middleware (อัตโนมัติ)

```go
router := gin.Default()

// เพิ่ม middleware เพื่อเก็บ metrics อัตโนมัติ
router.Use(metrics.HTTPMetricsMiddleware()) // วัด HTTP requests
router.Use(metrics.ErrorMetricsMiddleware()) // วัด errors
```

### 3. สร้าง Custom Metrics

```go
// นับจำนวนผู้ใช้ที่สมัคร
userCounter := metrics.NewCounter("user_registrations_total", "จำนวนการสมัครสมาชิก", metrics.Labels{})

// จำนวนผู้ใช้ออนไลน์
onlineUsers := metrics.NewGauge("users_online", "ผู้ใช้ออนไลน์", metrics.Labels{})

// เวลาในการประมวลผลออเดอร์
orderProcessTime := metrics.NewHistogram("order_process_time_seconds", "เวลาประมวลผลออเดอร์", nil, metrics.Labels{})
```

### 4. ใช้งานใน API

```go
// API สมัครสมาชิก
router.POST("/register", func(c *gin.Context) {
    // ... logic การสมัคร ...

    if success {
        // บันทึกว่ามีการสมัครสำเร็จ
        userCounter.With(metrics.Labels{
            "user_type": "premium",  // แยกตามประเภทผู้ใช้
            "source": "mobile",      // แยกตามแหล่งที่มา
        }).Inc()

        // เพิ่มจำนวนผู้ใช้ออนไลน์
        onlineUsers.Inc()
    }
})

// API ประมวลผลออเดอร์
router.POST("/orders", func(c *gin.Context) {
    // วัดเวลาการประมวลผล
    timer := metrics.NewTimer(orderProcessTime)
    stop := timer.Start()
    defer stop()

    // ... logic การประมวลผลออเดอร์ ...
})
```

### 5. ดูผลลัพธ์

```go
// API สำหรับดู metrics
router.GET("/metrics", func(c *gin.Context) {
    allMetrics := metrics.GetAllMetrics()
    c.JSON(200, gin.H{"metrics": allMetrics})
})
```

## 🏥 Health Checks - ตรวจสุขภาพระบบ

```go
// สร้าง health checker สำหรับ database
dbChecker := metrics.NewDatabaseHealthChecker("database", dbConnection)
metrics.RegisterHealthCheck(dbChecker)

// สร้าง health checker สำหรับ Redis
redisChecker := metrics.NewRedisHealthChecker("cache", redisClient)
metrics.RegisterHealthCheck(redisChecker)

// API ดูสุขภาพระบบ
router.GET("/health", func(c *gin.Context) {
    health, _ := metrics.GetOverallHealth(context.Background())
    c.JSON(200, gin.H{"status": health})
})
```

## 📊 ตัวอย่างผลลัพธ์ที่ได้

```json
{
  "metrics": [
    {
      "name": "myshop_api_http_requests_total",
      "type": "counter",
      "value": 1250,
      "labels": {
        "method": "GET",
        "status": "200",
        "endpoint": "/api/products"
      }
    },
    {
      "name": "myshop_api_users_online",
      "type": "gauge",
      "value": 45
    },
    {
      "name": "myshop_api_order_process_time_seconds",
      "type": "histogram",
      "value": 2.5 // ค่าเฉลี่ย
    }
  ]
}
```

## 🎯 ตัวอย่างการใช้งานในธุรกิจ

### ร้านค้าออนไลน์:

```go
// Business Metrics
orderCounter := metrics.NewCounter("orders_total", "จำนวนออเดอร์", labels)
revenueGauge := metrics.NewGauge("revenue_total", "ยอดขายรวม", labels)
productViews := metrics.NewCounter("product_views_total", "จำนวนการดูสินค้า", labels)

// เมื่อมีออเดอร์
func processOrder(amount float64, productID string) {
    orderCounter.With(metrics.Labels{
        "payment_method": "credit_card",
    }).Inc()

    revenueGauge.Add(amount)
}

// เมื่อมีคนดูสินค้า
func viewProduct(productID string) {
    productViews.With(metrics.Labels{
        "product_id": productID,
        "category": "electronics",
    }).Inc()
}
```

### แอปสื่อสังคม:

```go
// Social Media Metrics
postLikes := metrics.NewCounter("post_likes_total", "จำนวนไลค์", labels)
activeUsers := metrics.NewGauge("active_users", "ผู้ใช้ออนไลน์", labels)
feedLoadTime := metrics.NewHistogram("feed_load_time_seconds", "เวลาโหลด feed", nil, labels)

// เมื่อมีคนไลค์
func likePost(postID, userID string) {
    postLikes.With(metrics.Labels{
        "post_type": "photo",
        "user_type": "premium",
    }).Inc()
}
```

## 🚨 เคล็ดลับและข้อควรระวัง

### ✅ DO:

1. **ใช้ชื่อที่ชัดเจน** - `user_registrations_total` ดีกว่า `users`
2. **ใส่หน่วย** - `response_time_seconds` ดีกว่า `response_time`
3. **ใช้ label อย่างสมเหตุผล** - แยกข้อมูลที่เป็นประโยชน์
4. **เก็บข้อมูลที่เกี่ยวข้อง** - วัดสิ่งที่จะนำไปใช้ตัดสินใจ

### ❌ DON'T:

1. **Label เยอะเกินไป** - อย่าใส่ user_id เป็น label (จะมีเยอะมาก)
2. **ชื่อยาวเกินไป** - ใช้ชื่อที่เข้าใจง่าย
3. **วัดทุกอย่าง** - วัดแค่สิ่งที่สำคัญ
4. **ลืมใส่หน่วย** - อย่าลืมบอกว่าเป็นวินาที, ไบต์, หรืออะไร

## 🔍 การดู Metrics ในโลกจริง

### 1. ในแอป (API Endpoint):

```bash
curl http://localhost:8080/metrics
```

### 2. ใน Log:

```bash
tail -f logs/app.log | grep metrics
```

### 3. ในระบบ Monitoring (เช่น Grafana, Prometheus):

```
# แสดงกราฟ response time
rate(http_response_time_seconds[5m])

# แสดงจำนวน error
increase(http_requests_total{status=~"5.."}[1h])
```

## 🎯 สรุป: เริ่มต้นใช้ Metrics

1. **เพิ่ม HTTP Middleware** - ได้ metrics พื้นฐานทันที
2. **เพิ่ม Health Checks** - รู้ว่าระบบพร้อมใช้งานหรือไม่
3. **วัดสิ่งสำคัญ** - เช่น จำนวน user, เวลาตอบสนอง, ยอดขาย
4. **ดูผลลัพธ์** - สร้าง API `/metrics` และ `/health`
5. **ปรับปรุงเรื่อย ๆ** - เพิ่ม metrics ใหม่ตามความต้องการ

**เริ่มแค่นี้ก่อน แล้วค่อย ๆ เพิ่มไปเรื่อย ๆ!** 🚀
