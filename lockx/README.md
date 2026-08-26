# lockx

`lockx` 提供仅在单个 Go 进程内有效的按键互斥锁，适合保护同一实例内的重复提交、同一资源并发修改等场景。

它不提供分布式锁能力，不能用于跨进程、跨容器或跨机器的互斥。

## 使用

```go
import (
	"context"
	"errors"

	"github.com/jackman0925/go-foundation/lockx"
)

type OrderService struct {
	locker *lockx.KeyedLocker
}

func NewOrderService() *OrderService {
	return &OrderService{
		// 锁实例应随应用或服务对象一起长期持有。
		locker: lockx.NewKeyedLocker(),
	}
}

func (s *OrderService) Submit(ctx context.Context, orderID string) error {
	// 为同一订单构造稳定的锁键；不同订单可并行处理。
	key := "order:" + orderID
	unlock, err := s.locker.Lock(ctx, key)
	if err != nil {
		return err
	}
	defer unlock()

	// TODO: 查询订单当前状态，判断是否已经提交或已经处理完成。
	// TODO: 执行订单提交、库存扣减等需要同一订单串行处理的操作。
	// TODO: 保存处理结果。
	return nil
}
```

需要立即返回时使用 `TryLock`：

```go
func (s *OrderService) SubmitNow(ctx context.Context, orderID string) error {
	unlock, ok := s.locker.TryLock("order:" + orderID)
	if !ok {
		return errors.New("订单正在处理，请稍后重试")
	}
	defer unlock()

	// TODO: 执行不允许排队等待的订单处理逻辑。
	return nil
}
```

## 行为边界

- 同一 `KeyedLocker` 实例内，同一个 key 同时只会被一个调用方持有；不同 key 不互相阻塞。
- `Lock` 支持 `context.Context`，取消等待后会自动移除等待项。
- 返回的 `UnlockFunc` 可重复调用，但只会实际释放一次。
- 解锁后没有等待者的 key 会被自动清理，不会因历史 key 持续占用内存。
- 不支持重入、超时自动释放、租约续期和公平性承诺。
- 不要创建全局默认实例；由业务应用按生命周期显式持有 `KeyedLocker`。
