# Plano de Implementação: F06. Interactive Search & Replace System

**Pré-requisitos:**
- Go 1.22 ou superior instalado no ambiente de desenvolvimento
- Módulos `internal/buffer` (F01), `internal/theme` (F02), `internal/vim` (F03) e `internal/ui` (F07) existentes
- Motor de busca isolado em `internal/search` já compilando e testado

---

### Fase 1: Extensões na Máquina de Estados Modal do Vim (`internal/vim`)

**1. Estado Modal e Modos de Busca** - Expandir `internal/vim/state.go` adicionando a constante `ModeSearch`, suporte a `"SEARCH"` em `Mode.String()`, e os campos `SearchQuery`, `SearchIsReverse` e `SearchInitPos` na struct `ModalState`.

**2. Despachante de Teclas de Busca e Navegação** - Atualizar `internal/vim/engine.go` para capturar as teclas `/` e `?` em modo Normal (transicionando para `ModeSearch` e gravando posição inicial), processar teclas em `ModeSearch` (inserção de runes, Backspace, confirmação com Enter e cancelamento com restauração de cursor em Esc), e tratar `n` e `N` em modo Normal para saltos entre ocorrências.

**3. Comandos Ex de Substituição** - Integrar `internal/vim/command.go` com `search.ParseReplaceCommand` para que instruções `:s/.../.../` e `:%s/.../.../` executem substituições atômicas no buffer com snapshot único na pilha de histórico de Undo.

---

### Fase 2: Renderização de Prompt de Busca (`internal/ui`)

**4. Prompt de Busca na Linha de Comando** - Atualizar `internal/ui/statusline.go` para que `RenderCommandLine` renderize o prefixo `/` ou `?` seguido da query digitada quando o modo ativo for `ModeSearch`.

---

### Fase 3: Orquestração de Busca Incremental no App Model (`internal/app`)

**5. Integração de Busca em Tempo Real no Update** - Conectar os eventos de digitação em `ModeSearch` no `internal/app/model.go` para disparar `SearchEngine.Search` a cada alteração, rolar o viewport suavemente e posicionar o cursor no primeiro match encontrado.

**6. Navegação entre Ocorrências e Notificação de Wrap-Around** - Processar comandos de salto `n` e `N` invocando `SearchEngine.Next` e `SearchEngine.Prev`, posicionando o cursor na linha e coluna da ocorrência encontrada e exibindo avisos de wrap-around ou erro `E486: Padrão não encontrado`.

---

### Fase 4: Testes Automatizados e Validação

**7. Testes de Unidade e Integração** - Implementar casos de teste cobrindo transições de modo em `internal/vim/engine_test.go`, renderização do prompt em `internal/ui/statusline_test.go`, e o fluxo completo de busca incremental e navegação em `internal/app/model_vim_test.go`.
