package main

import (
	"flag"
	"os"
	"strconv"
)

//(4.1) Доработайте код, чтобы он умел принимать аргументы с использованием флагов.
//
//Аргументы агента:
//•	Флаг -a=<ЗНАЧЕНИЕ> отвечает за адрес эндпоинта HTTP-сервера (по умолчанию localhost:8080).
//•	Флаг -r=<ЗНАЧЕНИЕ> позволяет переопределять reportInterval — частоту отправки метрик на сервер (по умолчанию 10 секунд).
//•	Флаг -p=<ЗНАЧЕНИЕ> позволяет переопределять pollInterval — частоту опроса метрик из пакета runtime (по умолчанию 2 секунды).
//При попытке передать приложению незвестные флаги оно должно завершаться с сообщением о соответствующей ошибке.
//Значения интервалов времени должны задаваться в секундах.
//Во всех случаях должны присутствовать значения по умолчанию.

//(5.1) Доработайте агент, чтобы он мог изменять свои параметры запуска по умолчанию через переменные окружения:
//•	ADDRESS отвечает за адрес эндпоинта HTTP-сервера.
//•	REPORT_INTERVAL позволяет переопределять reportInterval (по умолчанию 10 секунд).
//•	POLL_INTERVAL позволяет переопределять pollInterval (по умолчанию 2 секунды).
//Значения интервалов времени должны задаваться в секундах.
//
//Приоритет параметров должен быть таким:
//•	Если указана переменная окружения, то используется она.
//•	Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
//•	Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.

// основное хранилище настроек агента
type AgentConfig struct {
	Endpoint       string
	ReportInterval int64
	PollInterval   int64
}

func InitAgentConfig() AgentConfig {

	cf := AgentConfig{}

	//Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
	flag.StringVar(&cf.Endpoint, "a", "localhost:8080", "the address:port server endpoint to send metric data")
	flag.Int64Var(&cf.ReportInterval, "r", 10, "metric reporting frequency in seconds")
	flag.Int64Var(&cf.PollInterval, "p", 2, "metric poll interval in seconds")
	flag.Parse()

	//Разбор переменных окружения
	if val, found := os.LookupEnv("ADDRESS"); found && (val != "") {
		cf.Endpoint = val
	}
	if val, found := os.LookupEnv("REPORT_INTERVAL"); found && (val != "") {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.ReportInterval = intval
		}
	}
	if val, found := os.LookupEnv("POLL_INTERVAL"); found && (val != "") {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.PollInterval = intval
		}
	}

	// // Access and print non-flag arguments
	// args := flag.Args()
	// fmt.Println("Non-flag arguments:", args)

	return cf
}
