# Especificação de Casos de Teste: F01. CLI Entrypoint & File Buffer Manager

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 7 | Manipulação de `Line`, runes UTF-8, estado de `Buffer` e dirty tracking |
| Integração | 6 | I/O de disco, gravação atômica, leitura de stdin, watcher `fsnotify` e contratos de saída |
| E2E / CLI | 4 | Invocação de linha de comando (`mdn`), pipes e encerramento com código de saída |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| O comando `mdn arquivo.md` abre o arquivo existente carregando seu conteúdo integral no buffer | TC04, TC13 |
| O comando `mdn novo.md` inicializa um buffer vazio exibindo indicador `[Novo Arquivo]` | TC05, TC14 |
| O comando `mdn` sem argumentos inicializa um buffer scratchpad vazio | TC02, TC15 |
| O comando `cat doc.md | mdn` carrega a saída completa do pipe no buffer inicial | TC06, TC16 |
| O comando `:w` grava o buffer no disco de forma atômica mantendo permissões e encoding UTF-8 | TC07, TC08, TC09 |
| O comando `:q` bloqueia a saída e emite mensagem de erro caso existam alterações não salvas | TC10 |
| O comando `:q!` força a saída imediata sem gravar alterações não salvas | TC11 |
| Cross-Feature: Buffer State fornece dados íntegros para outros módulos | TC01, TC03, TC12 |

---

## Casos de Teste Unitários

### TC01 — Manipulação de Linha com Caracteres UTF-8 Multibyte
- **Alvo:** `internal/buffer/line.go` (`NewLine`, `InsertAt`, `DeleteAt`)
- **Tipo:** unit
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Instanciar uma `Line` com a string contendo acentos e emojis: `"Olá, Mundo! 🚀\n"`.
  2. Verificar contagem de runes (`Length()`).
  3. Inserir o rune `'✨'` no índice 4.
  4. Deletar o rune na posição do emoji de foguete.
- **Resultado Esperado:** A contagem de runes reflete o número real de caracteres Unicode (e não contagem bruta de bytes), a inserção e deleção preservam os offsets corretos e `String()` reconverte para UTF-8 válido.
- **Cobre:** Manipulação correta de runes UTF-8 sem corrupção de caracteres multibyte.

### TC02 — Inicialização de Buffer Scratchpad Vazio
- **Alvo:** `internal/buffer/buffer.go` (`NewScratchpadBuffer`)
- **Tipo:** unit
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Chamar `NewScratchpadBuffer()`.
  2. Inspecionar `FilePath`, `IsDirty`, `IsNewFile`, `IsStdinBuffer` e `Lines`.
- **Resultado Esperado:** `FilePath` deve ser vazio, `IsDirty == false`, `IsNewFile == true`, `IsStdinBuffer == false` e o buffer deve conter exatamente 1 linha vazia com cursor em `(0, 0)`.
- **Cobre:** Critério PRD: `mdn` sem argumentos inicializa scratchpad vazio.

### TC03 — Rastreamento de Estado Modificado (Dirty Tracking)
- **Alvo:** `internal/buffer/buffer.go` (`InsertText`, `SetDirty`, `ResetDirty`)
- **Tipo:** unit
- **Pré-condições:** Buffer instanciado limpo (`IsDirty == false`).
- **Passos:**
  1. Inserir caractere ou texto no buffer.
  2. Verificar valor de `IsDirty`.
  3. Invocar `SetDirty(false)` (simulando pós-salvamento).
  4. Verificar valor de `IsDirty`.
- **Resultado Esperado:** `IsDirty` transiciona para `true` imediatamente após qualquer mutação de conteúdo e retorna para `false` após reset.
- **Cobre:** Integridade do estado de modificação do buffer.

### TC04 — Leitura de Arquivo Existente e Detecção de Quebra de Linha
- **Alvo:** `internal/buffer/io.go` (`LoadFromFile`)
- **Tipo:** unit / integração I/O
- **Pré-condições:** Arquivo temporário criado no disco com 3 linhas contendo terminador `\r\n` (CRLF) e acentuação em UTF-8.
- **Passos:**
  1. Invocar `LoadFromFile(path)`.
  2. Verificar `LineCount()`, conteúdo de cada linha e o tipo `Ending` detectado.
  3. Validar `IsNewFile == false` e `IsDirty == false`.
- **Resultado Esperado:** Todas as linhas são carregadas fielmente, o terminador CRLF é preservado e os metadados de arquivo (`OriginalModTime`, `OriginalMode`) são populados.
- **Cobre:** Critério PRD: `mdn arquivo.md` carrega conteúdo integral preservando line endings.

### TC05 — Abertura de Caminho Inexistente como Novo Arquivo
- **Alvo:** `internal/buffer/io.go` (`LoadFromFile`)
- **Tipo:** unit / integração I/O
- **Pré-condições:** Caminho de arquivo que não existe no disco (`/tmp/teste_novo_doc_123.md`).
- **Passos:**
  1. Invocar `LoadFromFile(caminhoInexistente)`.
  2. Verificar se nenhum erro fatal é retornado.
  3. Inspecionar campos do buffer retornado.
- **Resultado Esperado:** Buffer inicializado com 1 linha vazia, `FilePath` igual ao caminho informado, `IsNewFile == true` e `IsDirty == false`.
- **Cobre:** Critério PRD: `mdn novo.md` inicializa novo buffer vazio.

### TC06 — Leitura de Dados via Reader / Pipe Stdin
- **Alvo:** `internal/buffer/io.go` (`LoadFromReader`)
- **Tipo:** unit
- **Pré-condições:** Stream `io.Reader` fornecido com conteúdo Markdown simulando pipe CLI.
- **Passos:**
  1. Invocar `LoadFromReader(strings.NewReader("# Documento Pipe\nLinha 2\n"), "[Stdin Buffer]")`.
  2. Verificar `Lines`, `IsStdinBuffer`, `FilePath` e `IsDirty`.
- **Resultado Esperado:** `LineCount() == 2`, `IsStdinBuffer == true`, `FilePath == "[Stdin Buffer]"` e `IsDirty == false`.
- **Cobre:** Critério PRD: `cat doc.md | mdn` carrega saída do pipe no buffer inicial.

### TC07 — Validação de Erro ao Salvar Buffer sem Caminho
- **Alvo:** `internal/buffer/io.go` (`AtomicSave`)
- **Tipo:** unit
- **Pré-condições:** Buffer scratchpad criado sem `FilePath`.
- **Passos:**
  1. Invocar `AtomicSave(buf, "")`.
- **Resultado Esperado:** Retorna erro explícito com mensagem: `Erro: Nenhum nome de arquivo definido. Use :w <caminho>`.
- **Cobre:** Tratamento de erro PRD: tentativa de salvar buffer sem nome.

---

## Casos de Teste de Integração

### TC08 — Gravação Atômica e Preservação de Permissões
- **Alvo:** `internal/buffer/io.go` (`AtomicSave`)
- **Tipo:** integration
- **Pré-condições:** Arquivo existente criado com permissão `0755` (executável) e conteúdo inicial.
- **Passos:**
  1. Carregar o arquivo para o buffer e alterar o conteúdo de uma linha.
  2. Executar `AtomicSave(buf, path)`.
  3. Inspecionar o arquivo em disco pós-salvamento: ler bytes e verificar `os.Stat(path).Mode()`.
- **Resultado Esperado:** O arquivo é gravado atomicamente com o novo conteúdo, sem deixar arquivos temporários residuais `.tmp.*` no diretório e mantendo as permissões `0755`.
- **Cobre:** Critério PRD: `:w` grava buffer de forma atômica mantendo permissões e encoding.

### TC09 — Criação Automática de Diretórios Pais ao Salvar (Full Scope)
- **Alvo:** `internal/buffer/io.go` (`AtomicSave`)
- **Tipo:** integration
- **Pré-condições:** Caminho de destino com diretórios aninhados ainda não existentes (`/tmp/sandbox/sub1/sub2/novo.md`).
- **Passos:**
  1. Criar buffer com o caminho de destino aninhado.
  2. Executar `AtomicSave(buf, targetPath)`.
  3. Verificar se as pastas intermediárias foram criadas no disco e o arquivo existe.
- **Resultado Esperado:** Diretórios `/tmp/sandbox/sub1/sub2/` criados com sucesso (`0755`) e arquivo salvo com conteúdo integral.
- **Cobre:** Full Scope PRD: criação automática de diretórios pais inexistentes ao salvar.

### TC10 — Bloqueio de Saída com Alterações Não Salvas (`:q`)
- **Alvo:** `internal/app/model.go` (`Update` com `QuitMsg{Force: false}`)
- **Tipo:** integration
- **Pré-condições:** Modelo Bubble Tea instanciado com buffer modificado (`IsDirty == true`).
- **Passos:**
  1. Despachar comando de saída sem force (`QuitMsg{Force: false}`).
  2. Avaliar retorno de `Update`.
- **Resultado Esperado:** O comando não emite `tea.Quit`, o editor permanece em execução e define a mensagem de erro na barra de status: `E37: Alterações não salvas. Use :w para salvar ou :q! para forçar a saída`.
- **Cobre:** Critério PRD: `:q` bloqueia saída e emite mensagem de erro caso existam alterações pendentes.

### TC11 — Forçar Saída com Alterações Pendentes (`:q!`)
- **Alvo:** `internal/app/model.go` (`Update` com `QuitMsg{Force: true}`)
- **Tipo:** integration
- **Pré-condições:** Modelo Bubble Tea instanciado com buffer modificado (`IsDirty == true`).
- **Passos:**
  1. Despachar comando de saída forçada (`QuitMsg{Force: true}`).
  2. Avaliar comando retornado por `Update`.
- **Resultado Esperado:** `Update` retorna imediatamente o comando `tea.Quit`, permitindo encerramento do processo sem salvar alterações.
- **Cobre:** Critério PRD: `:q!` força a saída imediata sem gravar alterações não salvas.

### TC12 — Monitoramento Externo e Auto-Recarregamento com fsnotify (Full Scope)
- **Alvo:** `internal/watcher/watcher.go` & `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Arquivo aberto no buffer limpo (`IsDirty == false`) com watcher ativo.
- **Passos:**
  1. Escrever externamente novas linhas no arquivo em disco usando `os.WriteFile`.
  2. Aguardar evento `FileModifiedMsg` processado pelo `Update` do modelo.
  3. Verificar o conteúdo das linhas no buffer.
- **Resultado Esperado:** O buffer detecta a alteração externa e recarrega as novas linhas automaticamente sem exigir intervenção manual. Se o buffer estivesse com `IsDirty == true`, não sobrescreveria e emitiria aviso informativo.
- **Cobre:** Full Scope PRD: recarregamento automático de arquivo modificado externamente.

---

## Casos de Teste E2E / CLI

### TC13 — Invocação CLI com Arquivo Existente
- **Alvo:** Executável compilado `mdn`
- **Tipo:** e2e
- **Pré-condições:** Arquivo `teste.md` no disco com texto `"# Header"`.
- **Passos:**
  1. Executar `./mdn teste.md` em subprocesso com timeout.
  2. Enviar sequência de saída `:q\n`.
- **Resultado Esperado:** Processo inicializa com sucesso e encerra com código 0.
- **Cobre:** Critério PRD: comando `mdn arquivo.md` abre arquivo existente.

### TC14 — Invocação CLI com Novo Arquivo
- **Alvo:** Executável compilado `mdn`
- **Tipo:** e2e
- **Pré-condições:** Nenhum arquivo `nao_existe.md`.
- **Passos:**
  1. Executar `./mdn nao_existe.md`.
  2. Enviar sequência de saída `:q\n`.
- **Resultado Esperado:** Processo inicializa com indicador de novo arquivo e encerra com código 0.
- **Cobre:** Critério PRD: comando `mdn novo.md` inicializa novo buffer.

### TC15 — Invocação CLI sem Argumentos (Scratchpad)
- **Alvo:** Executável compilado `mdn`
- **Tipo:** e2e
- **Pré-condições:** Nenhuma.
- **Passos:**
  1. Executar `./mdn` sem argumentos.
  2. Enviar sequência de saída `:q\n`.
- **Resultado Esperado:** Processo abre buffer scratchpad e encerra com código 0.
- **Cobre:** Critério PRD: comando `mdn` sem argumentos inicializa scratchpad vazio.

### TC16 — Invocação CLI via Pipe Stdin
- **Alvo:** Executável compilado `mdn`
- **Tipo:** e2e
- **Pré-condições:** Stream de texto preparado.
- **Passos:**
  1. Executar `echo "texto via pipe" | ./mdn`.
  2. Enviar sequência de saída `:q\n`.
- **Resultado Esperado:** Processo consome os dados do pipe, inicializa o buffer como `[Stdin Buffer]` e encerra com código 0.
- **Cobre:** Critério PRD: `cat doc.md | mdn` carrega saída do pipe no buffer inicial.
