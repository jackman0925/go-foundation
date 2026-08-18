# timex

`timex` 提供常用时间格式化、解析、边界计算、时间戳转换、日期比较和星期工具。

## 导入

```go
import "github.com/jackman0925/go-foundation/timex"
```

## 基础用法

```go
start := timex.StartOfDay(time.Now())
end := timex.EndOfDay(time.Now())
text := timex.FormatDateTime(time.Now())
```

## 常用格式

```go
date := timex.FormatDate(time.Now())                  // 2026-08-18
datetime := timex.FormatDateTime(time.Now())          // 2026-08-18 09:10:11
compact := timex.FormatCompactDateTime(time.Now())    // 20260818091011
millis := timex.FormatCompactDateTimeMilli(time.Now())       // 20260818091011123
shortMS := timex.FormatShortCompactDateTimeMilli(time.Now()) // 260818091011123
```

## 解析

```go
t, err := timex.ParseTime("2026-08-18 09:10:11")
t, err = timex.ParseTime("2026/08/18 09:10:11")
t, err = timex.ParseTime("20260818091011")
t, err = timex.ParseTimeInLocation("2026-08-18", time.Local)
```

## 边界和日期计算

```go
dayStart := timex.StartOfDay(t)
dayEnd := timex.EndOfDay(t)
weekStart := timex.StartOfWeek(t)
monthEnd := timex.EndOfMonth(t)

nextDay := timex.AddDays(t, 1)
nextMonth := timex.AddMonths(t, 1)
```

## 时间戳和比较

```go
millis := timex.UnixMilli(t)
t = timex.FromUnixMilli(millis)

result, err := timex.CompareDateString("2026-08-18", "2026-08-19")
seconds := timex.SecondsBetween(start, end)
```

## 星期和年龄

```go
name := timex.ChineseWeekday(time.Now()) // 星期二
value := timex.ISOWeekday(time.Now()) // 周一为 1，周日为 7
age := timex.Age(birthday)
```

## 注意事项

- 日、周、月边界会保留输入时间的 location；
- `Parse` 支持横杠、斜杠、紧凑日期时间和 RFC3339 格式；
- `CompareDateString` 只比较 `2006-01-02` 格式的日期字符串；
- `AgeAt` 适合单元测试固定当前时间，业务代码可直接使用 `Age`。
