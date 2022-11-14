package siid

import (
	"context"
	"errors"
)

var (
	ErrReachIdLimitation    = errors.New("reach id limit")
	ErrIdRunOut             = errors.New("id run out")
	ErrorDriverHasClosed    = errors.New("driver has closed")
	ErrorDriverHasNotInited = errors.New("driver has not inited, call Builder.Prepare first")
)

type Stats struct {
	Current       uint64 // 当前id值
	Max           uint64 // id最大值
	RenewCount    uint64 // renew的次数
	RenewErrCount uint64 // renew的错误次数，若>0，属于发生了严重错误
}

type Builder interface {
	// Prepare 负责准备工作，会调用Driver.Prepare函数
	Prepare(context.Context) error

	// Destroy 销毁，资源的释放，会调用Driver.Destroy函数
	Destroy(context.Context) error

	// Build 建立Engine（新建或者返回已存在的Engine）
	// domain 域，每种类型id，都拥有一个固定的域名，例如`player`
	Build(domain string) (Engine, error)
}

type Engine interface {
	// Next 取唯一id
	Next() (uint64, error)

	// MustNext 取唯一id，若发生错误，则会panic
	MustNext() uint64

	// Stats 当前状态
	Stats() Stats
}

type Driver interface {
	// Prepare 负责准备工作
	Prepare(context.Context) error

	// Destroy 销毁，资源的释放
	Destroy(context.Context) error

	// Renew 创建新的一段id
	// domain 域，每种类型id，都拥有一个固定的域名，例如`player`
	// quantum 段长，创建一段id的段长
	// offsetOnCreate 若为新域，偏移多少开始创建
	// 返回当前id
	Renew(ctx context.Context, domain string, quantum, offsetOnCreate uint64) (uint64, error)
}
