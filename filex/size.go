package filex

import "strconv"

// FormatDecimalSize 将字节数格式化为十进制文本，最大单位为 PB，仅用于展示。
func FormatDecimalSize(bytes int64) string {
	return formatSize(bytes, 1000, [...]string{"B", "kB", "MB", "GB", "TB", "PB"})
}

// FormatBinarySize 将字节数格式化为 IEC 二进制文本，最大单位为 PiB，仅用于展示。
func FormatBinarySize(bytes int64) string {
	return formatSize(bytes, 1024, [...]string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"})
}

func formatSize(bytes int64, base uint64, units [6]string) string {
	sign := ""
	value := uint64(bytes)
	if bytes < 0 {
		sign = "-"
		// 避免直接对 MinInt64 取负导致溢出。
		value = uint64(-(bytes + 1)) + 1
	}
	unit, divisor := 0, uint64(1)
	for unit < len(units)-1 && value/divisor >= base {
		unit++
		divisor *= base
	}
	if unit == 0 {
		return sign + strconv.FormatUint(value, 10) + " B"
	}
	for {
		scale := uint64(10)
		if unit >= 3 {
			scale = 100
		}
		// 分离整数与余数，避免完整字节数乘以精度系数溢出。
		// 最大除数为 1024^5，余数乘 100 仍在 uint64 范围内。
		rounded := value/divisor*scale + (value%divisor*scale+divisor/2)/divisor
		if rounded >= base*scale && unit < len(units)-1 {
			unit++
			divisor *= base
			continue
		}
		text := strconv.FormatUint(rounded/scale, 10)
		if fraction := rounded % scale; fraction != 0 {
			if scale == 100 && fraction%10 == 0 {
				text += "." + strconv.FormatUint(fraction/10, 10)
			} else if scale == 100 && fraction < 10 {
				text += ".0" + strconv.FormatUint(fraction, 10)
			} else {
				text += "." + strconv.FormatUint(fraction, 10)
			}
		}
		return sign + text + " " + units[unit]
	}
}
