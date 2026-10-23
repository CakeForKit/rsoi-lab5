package utils

import "strings"

func SubstringLast(value string, str string) (string, string) {
	//возвращает индекс последней найденной подстроки в строке
	pos := strings.LastIndex(value, str)
	if pos == -1 {
		return value, ""
	}
	//доавляем к этому индексу длину подстроки (получаем позицию, где она кончается)
	adjustedPos := pos + len(str)

	//если больше или равно возвращаем все, что идет до
	if adjustedPos >= len(value) {
		return value[:pos], ""
	}
	//если нет, возвращаем все то, что идет до и после
	return value[:pos], value[adjustedPos:]
}
