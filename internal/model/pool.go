package model

import "sync"

// Resetter — интерфейс типов, поддерживающих сброс состояния (для переиспользования в пуле).
type Resetter interface {
	Reset()
}

// Pool — контейнер для переиспользования объектов типа T с методом Reset().
// Перед возвратом в пул объект сбрасывается вызовом Reset().
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New создаёт и возвращает указатель на пул для типа T.
// Параметр newFunc вызывается при Get(), когда пул пуст — так создаётся новый объект.
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return newFunc() },
		},
	}
}

// Get возвращает объект из пула. Если пул пуст, создаётся новый объект через функцию, переданную в New.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put помещает объект в пул. Перед помещением вызывается Reset() для сброса состояния объекта.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
