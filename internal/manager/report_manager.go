package manager

import (
	"encoding/json"
	"fmt"
	l "github.com/faelmori/logz"
	"os"
	"sync"
)

type ReportManager[T any] struct {
	notifyChan chan T              // Canal de notificação
	Files      map[string]*os.File // Map de tipos para arquivos abertos
	logger     l.Logger            // Logger para registrar eventos
	mu         sync.Mutex          // Mutex pra garantir operações seguras
}

// NewReportManager cria uma nova instância do ReportManager
func NewReportManager[T any](logger l.Logger) *ReportManager[T] {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &ReportManager[T]{
		mu:         sync.Mutex{},
		logger:     logger,
		notifyChan: make(chan T, 20),
		Files:      make(map[string]*os.File),
	}
}

func (rm *ReportManager[T]) Write(data *T, reportType string, format string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	file, exists := rm.Files[reportType]
	if !exists {
		var err error
		file, err = os.OpenFile(reportType+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("erro ao abrir/criar arquivo de report: %v", err)
		}
		rm.Files[reportType] = file
	}

	rm.logger.InfoCtx(fmt.Sprintf("Escrevendo report no arquivo %s", reportType), nil)

	// Escreve no formato especificado
	return GenerateReport(data, reportType+".log", format)
}

func GenerateReport[T any](data *T, filePath string, format string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("erro ao abrir/criar arquivo de report: %v", err), nil)
		return fmt.Errorf("erro ao criar arquivo de report: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Escrevendo report no arquivo %s", filePath), nil)

	// Escrever o report no formato definido
	switch format {
	default:
		//case "json":
		dataBytes, err := json.Marshal(data)
		if err != nil {
			l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("erro ao converter dados para JSON: %v", err), nil)
			return fmt.Errorf("erro ao converter dados para JSON: %v", err)
		}
		_, err = file.Write(dataBytes)
		if err != nil {
			l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("erro ao escrever dados no arquivo: %v", err), nil)
			return fmt.Errorf("erro ao escrever dados no arquivo: %v", err)
		}
		_, err = file.WriteString("\n")
		if err != nil {
			l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("erro ao escrever nova linha no arquivo: %v", err), nil)
			return fmt.Errorf("erro ao escrever nova linha no arquivo: %v", err)
		}
	}

	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Report escrito com sucesso no arquivo %s", filePath), nil)

	return nil
}

//
//func writeCSV[T any](file *os.File, data *T) error {
//	stringData, err := json.Marshal(data)
//	if err != nil {
//		return fmt.Errorf("erro ao converter dados para CSV: %v", err)
//	}
//	// Aqui você implementaria a lógica de conversão de JSON para CSV
//
//	return nil
//}
