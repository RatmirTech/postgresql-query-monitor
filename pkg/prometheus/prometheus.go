package prometheus

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricType описывает тип метрики
type MetricType int

// Перечисление типов метрик Prometheus.
const (
	Gauge     MetricType = iota // Gauge — метрика, отражающая текущее значение.
	Counter                     // Counter — счетчик, только увеличивается.
	Histogram                   // Histogram — распределение значений по корзинам.
)

// PrometheusManager централизованно управляет метриками
type PrometheusManager struct {
	mutex      sync.RWMutex
	gauges     map[string]*prometheus.GaugeVec
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
}

// NewManager создаёт новый PrometheusManager
func NewManager() *PrometheusManager {
	return &PrometheusManager{
		gauges:     make(map[string]*prometheus.GaugeVec),
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

// RegisterGauge регистрирует Gauge метрику с динамическими labels
func (m *PrometheusManager) RegisterGauge(name, help string, labels []string) *prometheus.GaugeVec {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if existing, ok := m.gauges[name]; ok {
		return existing
	}

	g := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)

	m.gauges[name] = g
	return g
}

// RegisterCounter регистрирует Counter метрику с labels
func (m *PrometheusManager) RegisterCounter(name, help string, labels []string) *prometheus.CounterVec {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if existing, ok := m.counters[name]; ok {
		return existing
	}

	c := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)

	m.counters[name] = c
	return c
}

// RegisterHistogram регистрирует Histogram метрику с labels
func (m *PrometheusManager) RegisterHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if existing, ok := m.histograms[name]; ok {
		return existing
	}

	h := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labels)

	m.histograms[name] = h
	return h
}

// SetGauge устанавливает значение Gauge с указанными label
func (m *PrometheusManager) SetGauge(name string, labels prometheus.Labels, value float64) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	g, ok := m.gauges[name]
	if !ok {
		return fmt.Errorf("gauge %s not registered", name)
	}

	g.With(labels).Set(value)
	return nil
}

// IncCounter увеличивает Counter с указанными label
func (m *PrometheusManager) IncCounter(name string, labels prometheus.Labels, delta float64) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	c, ok := m.counters[name]
	if !ok {
		return fmt.Errorf("counter %s not registered", name)
	}

	c.With(labels).Add(delta)
	return nil
}

// ObserveHistogram добавляет наблюдение в Histogram с label
func (m *PrometheusManager) ObserveHistogram(name string, labels prometheus.Labels, value float64) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	h, ok := m.histograms[name]
	if !ok {
		return fmt.Errorf("histogram %s not registered", name)
	}

	h.With(labels).Observe(value)
	return nil
}

// Handler возвращает http.Handler для отдачи всех метрик Prometheus
func Handler() http.Handler {
	return promhttp.Handler()
}
