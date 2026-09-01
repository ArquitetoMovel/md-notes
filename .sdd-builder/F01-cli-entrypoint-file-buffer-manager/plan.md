# Plano de Implementação: F01. CLI Entrypoint & File Buffer Manager

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Dependências Go declaradas em `go.mod`:
  - `github.com/charmbracelet/bubbletea` (v0.25+)
  - `github.com/fsnotify/fsnotify` (v1.7+)
  - `golang.org/x/term` (v0.18+)

---

### Fase 1: Fundação do Módulo Go e Estrutura de Dados do Buffer

**1. Módulo Go e Estrutura de Linha** - Inicializar o módulo `go.mod` do projeto e implementar o componente `internal/buffer/line.go` para manipulação de caracteres UTF-8 e controle de quebras de linha conforme especificado na arquitetura.

**2. Estrutura Central do Buffer** - Criar o componente `internal/buffer/buffer.go` com os campos de estado, controle de posição do cursor, metadados do arquivo e rastreamento de modificações não salvas.

**3. Operações de Mutação do Buffer** - Implementar as funções de inserção, remoção, leitura de linhas e reinicialização de flags de modificação no buffer para suportar operações de edição em memória.

---

### Fase 2: Mecanismos de I/O, Leitura de Stdin e Persistência Atômica

**4. Carregamento de Arquivos em Disco** - Implementar em `internal/buffer/io.go` a leitura de arquivos existentes em UTF-8 com detecção de terminações de linha e inicialização de buffers para novos arquivos.

**5. Leitura de Stream de Entrada Padrão (Stdin Pipe)** - Desenvolver no módulo de I/O a capacidade de ler fluxos completos de entrada via pipe até EOF e preencher o buffer scratchpad inicial.

**6. Persistência Atômica em Disco** - Construir o mecanismo de escrita segura em arquivo temporário com sincronização de disco, criação de pastas pais inexistentes e renomeação atômica mantendo as permissões de arquivo originais.

---

### Fase 3: Monitoramento Externo de Arquivos e Orquestração Bubble Tea

**7. Serviço de Monitoramento de Arquivos** - Implementar o componente `internal/watcher/watcher.go` utilizando eventos de sistema para detectar alterações externas em disco e despachar mensagens para o ciclo de eventos.

**8. Modelo Base da Aplicação Bubble Tea** - Criar em `internal/app/model.go` a estrutura de modelo Bubble Tea para integrar o ciclo de vida do buffer, recepção de eventos do watcher e processamento inicial de comandos.

**9. Controle de Saída e Prevenção de Perda de Dados** - Integrar no modelo da aplicação a regra de bloqueio de saída com aviso quando houver alterações pendentes e o suporte a encerramento forçado.

---

### Fase 4: Entrypoint CLI e Integração de Terminal TTY

**10. Ponto de Entrada do Executável** - Desenvolver `cmd/mdn/main.go` para processar parâmetros de linha de comando, identificar caminhos de arquivo ou iniciar buffers scratchpad.

**11. Detecção de Pipe e Reconexão de TTY** - Implementar no ponto de entrada a detecção de entrada por pipe e a reanexação do descritor de terminal interativo para a interface Bubble Tea.

**12. Tratamento de Erros Iniciais e Saída do Processo** - Configurar o tratamento de erros críticos de permissão no console e encerramento com código de status adequado antes do início da interface gráfica.
