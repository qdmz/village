import { themeColor } from './theme'

export const PALETTE = () => [themeColor(), '#84cc16', '#f59e0b', '#3b82f6', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#64748b']

export function pie(data, opts = {}) {
  return {
    color: PALETTE(),
    tooltip: { trigger: 'item', formatter: '{b}：{c}（{d}%）' },
    legend: { bottom: 0, icon: 'circle', itemWidth: 8, textStyle: { fontSize: 12 } },
    series: [{ type: 'pie', radius: opts.rose ? ['18%', '68%'] : ['46%', '70%'], roseType: opts.rose ? 'radius' : undefined, center: ['50%', '44%'],
      itemStyle: { borderColor: '#fff', borderWidth: 2, borderRadius: 6 }, label: { show: !!opts.label, formatter: '{b}\n{c}' },
      data: data.filter(d => d.value > 0) }]
  }
}

export function bar(names, series, opts = {}) {
  return {
    color: opts.color || PALETTE(),
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: series.length > 1 ? { top: 0, icon: 'roundRect', itemWidth: 10, itemHeight: 10 } : undefined,
    grid: { left: 8, right: 12, top: series.length > 1 ? 34 : 16, bottom: 4, containLabel: true },
    xAxis: opts.horizontal ? { type: 'value', splitLine: { lineStyle: { type: 'dashed' } } } : { type: 'category', data: names, axisTick: { show: false }, axisLabel: { interval: 0, fontSize: 11, rotate: opts.rotate || 0 } },
    yAxis: opts.horizontal ? { type: 'category', data: names, axisTick: { show: false }, inverse: true } : { type: 'value', name: opts.unit, splitLine: { lineStyle: { type: 'dashed' } } },
    series: series.map(s => ({ type: s.line ? 'line' : 'bar', smooth: true, barMaxWidth: 26, itemStyle: { borderRadius: opts.horizontal ? [0, 6, 6, 0] : [6, 6, 0, 0] }, areaStyle: s.area ? { opacity: 0.15 } : undefined, ...s }))
  }
}
