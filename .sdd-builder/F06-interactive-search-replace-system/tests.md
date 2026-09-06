# Especificação de Casos de Teste: F06. Interactive Search & Replace System

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 10 | Modos de busca, transições Vim, prompt UI, regex, smart-case e parser :s |
| Integração | 7 | Busca incremental em tempo real, navegação cíclica n/N, wrap-around, rollback com Esc, replace global |
| Performance | 2 | Varredura instantânea em 50k linhas (< 5ms) e busca incremental contínua |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| Pressionar `/` no modo Normal abre o prompt de busca no rodapé | TC01, TC10 |
| Digitar no prompt de busca destaca todas as ocorrências correspondentes no viewport em tempo real | TC02, TC11 |
| Pressionar `Enter` confirma a busca e posiciona o cursor no primeiro resultado | TC03, TC12 |
| Pressionar `n` move o cursor para a próxima ocorrência e `N` move para a ocorrência anterior | TC04, TC05, TC13 |
| O comando `:%s/antigo/novo/g` substitui todas as ocorrências no documento informando a contagem de substituições | TC07, TC14 |
| Buscar por um termo inexistente exibe mensagem de erro apropriada na barra de status | TC06, TC15 |
| Full Scope: Cancelamento com Esc, Busca reversa `?`, Smart-Case, wrap-around | TC08, TC09, TC16, TC17 |

---

## Casos de Teste Unitários

### TC01 — Abertura do Prompt de Busca com `/` no Vim Engine
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** `vim.Engine` em `ModeNormal`.
- **Passos:**
  1. Enviar mensagem de tecla `/`.
- **Resultado Esperado:** O modo ativo torna-se `ModeSearch`, `SearchIsReverse == false`, `SearchQuery == ""` e `SearchInitPos` armazena a posição prévia do cursor.
- **Covers:** Pressionar `/` abre prompt de busca.

### TC02 — Digitação e Edição no Modo de Busca
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** `vim.Engine` em `ModeSearch`.
- **Passos:**
  1. Digitar caracteres `"f"`, `"o"`, `"o"`.
  2. Pressionar `Backspace`.
- **Resultado Esperado:** `SearchQuery` passa por `"f"`, `"fo"`, `"foo"`, e após backspace torna-se `"fo"`.
- **Covers:** Digitação no prompt de busca.

### TC03 — Cancelamento de Busca com `Esc` Restaurando Cursor
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** `vim.Engine` em `ModeSearch` com cursor temporariamente deslocado para linha 5, coluna 2. Posição inicial gravada na linha 1, coluna 0.
- **Passos:**
  1. Pressionar `Esc`.
- **Resultado Esperado:** Modo retorna para `ModeNormal`, cursor é restaurado para linha 1, coluna 0, e a query de busca é limpa.
- **Covers:** Cancelamento seguro da busca.

### TC04 — Mapeamento das Teclas `n` e `N` no Modo Normal
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** `vim.Engine` em `ModeNormal` com última busca registrada.
- **Passos:**
  1. Pressionar tecla `n`.
  2. Pressionar tecla `N`.
- **Resultado Esperado:** Retorna comandos de ação de navegação para frente e para trás, respectivamente.
- **Covers:** Teclas `n` e `N` movem para próxima e anterior ocorrência.

### TC05 — Renderização do Prompt de Busca na Linha de Comando
- **Alvo:** `internal/ui/statusline.go` (`RenderCommandLine`)
- **Tipo:** unit
- **Pré-condições:** `StatusState` com `Mode: ModeSearch`, `CommandInput: "teste"`.
- **Passos:**
  1. Invocar `RenderCommandLine(state, theme, width)`.
- **Resultado Esperado:** A string gerada contém o prefixo `"/teste"`.
- **Covers:** Visualização do prompt de busca inferior.

### TC06 — Formatação de Erro de Padrão Não Encontrado `E486`
- **Alvo:** `internal/ui/statusline.go` (`RenderCommandLine`)
- **Tipo:** unit
- **Pré-condições:** `StatusState` com `StatusMessage: "E486: Padrão não encontrado: xyz"`.
- **Passos:**
  1. Invocar `RenderCommandLine(state, theme, width)`.
- **Resultado Esperado:** Renderiza a mensagem em destaque de erro com código E486.
- **Covers:** Mensagem de erro apropriada para termo inexistente.

### TC07 — Análise Sintática do Comando Ex `:s` e `:%s`
- **Alvo:** `internal/search/replace.go` (`ParseReplaceCommand`)
- **Tipo:** unit
- **Pré-condições:** Strings de comando: `":%s/antigo/novo/g"`, `":s/foo/bar/g"`.
- **Passos:**
  1. Executar `ParseReplaceCommand` para cada instrução.
- **Resultado Esperado:** Identifica corretamente padrão, substituto e escopo (global ou linha).
- **Covers:** Parser de comandos ex de substituição.

### TC08 — Busca Reversa com `?`
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** `vim.Engine` em `ModeNormal`.
- **Passos:**
  1. Pressionar tecla `?`.
- **Resultado Esperado:** O modo passa para `ModeSearch` com `SearchIsReverse == true`.
- **Covers:** Busca reversa iniciada por `?`.

### TC09 — Regra de Sensibilidade Inteligente (*Smart-Case*)
- **Alvo:** `internal/search/pattern.go`
- **Tipo:** unit
- **Pré-condições:** String com variações de caixa `"palavra"`, `"Palavra"`.
- **Passos:**
  1. Compilar query `"palavra"` e buscar.
  2. Compilar query `"Palavra"` e buscar.
- **Resultado Esperado:** Query em minúsculas encontra ambas as ocorrências; query com maiúscula encontra somente `"Palavra"`.
- **Covers:** Suporte a smart-case.

### TC10 — Fallback Transparente para Busca Literal
- **Alvo:** `internal/search/pattern.go`
- **Tipo:** unit
- **Pré-condições:** Digitação de caractere especial incompleto `"func("`.
- **Passos:**
  1. Compilar padrão `"func("`.
- **Resultado Esperado:** Compilação não falha com erro de regex; ativa busca literal retornando intervalos corretos.
- **Covers:** Fallback literal para regex incompleto.

---

## Casos de Teste de Integração

### TC11 — Busca Incremental em Tempo Real no App Model
- **Alvo:** `internal/app/model.go` & `internal/search/engine.go`
- **Tipo:** integration
- **Pré-condições:** Buffer com texto contendo `"antigo"` nas linhas 2 e 8. Cursor na linha 0.
- **Passos:**
  1. Enviar tecla `/`.
  2. Enviar teclas `"a"`, `"n"`, `"t"`.
- **Resultado Esperado:** A cada tecla, `m.SearchEngine.Result.Matches` é atualizado em tempo real e o cursor salta para o primeiro match na linha 2.
- **Covers:** Busca incremental em tempo real.

### TC12 — Confirmação de Busca com `Enter`
- **Alvo:** `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Busca ativa com query `"antigo"` e 2 matches encontrados.
- **Passos:**
  1. Pressionar `Enter`.
- **Resultado Esperado:** Retorna ao modo Normal, cursor permanece sobre a primeira ocorrência, status line exibe `[1/2]` e os destaques continuam ativos.
- **Covers:** Confirmação de busca com Enter.

### TC13 — Navegação Cíclica com `n` e `N` no App Model
- **Alvo:** `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Busca confirmada com 2 matches: Linha 2 e Linha 8. Cursor na Linha 2.
- **Passos:**
  1. Pressionar `n` -> cursor salta para Linha 8, status exibe `[2/2]`.
  2. Pressionar `n` -> wrap-around ocorre, cursor retorna para Linha 2, notificação de ciclo é exibida.
  3. Pressionar `N` -> cursor retorna para Linha 8.
- **Resultado Esperado:** Navegação bidirecional e wrap-around funcionam com precisão.
- **Covers:** Navegação com n e N e wrap-around.

### TC14 — Substituição Global Atômica (`:%s/antigo/novo/g`)
- **Alvo:** `internal/vim/command.go` & `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Buffer com 10 ocorrências da palavra `"legado"`.
- **Passos:**
  1. Digitar `:%s/legado/moderno/g` e pressionar `Enter`.
  2. Verificar buffer e contagem de alterações.
  3. Pressionar `u` no modo Normal.
- **Resultado Esperado:** Todas as 10 ocorrências são substituídas em um único passo e revertidas atomicamente com um único toque em `u`.
- **Covers:** Substituição global atômica e integração com histórico de Undo.

### TC15 — Feedback para Termo Inexistente
- **Alvo:** `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Buffer sem a palavra `"inexistente"`.
- **Passos:**
  1. Digitar `/inexistente` e pressionar `Enter`.
- **Resultado Esperado:** Cursor retorna à posição inicial e a barra inferior exibe `E486: Padrão não encontrado: inexistente`.
- **Covers:** Erro apropriado para padrão inexistente.

### TC16 — Busca Reversa Navegando com `n`
- **Alvo:** `internal/app/model.go`
- **Tipo:** integration
- **Pré-condições:** Cursor na linha 10. Ocorrências nas linhas 3, 7 e 12.
- **Passos:**
  1. Iniciar com `?` e buscar termo.
  2. Pressionar `n`.
- **Resultado Esperado:** Cursor salta para a linha 7 (direção para cima/trás no documento).
- **Covers:** Busca reversa funcional.

### TC17 — Sessão Interativa de Confirmação (`gc`)
- **Alvo:** `internal/search/interactive.go`
- **Tipo:** integration
- **Pré-condições:** Buffer com 3 ocorrências. Comando `:%s/foo/bar/gc`.
- **Passos:**
  1. Responder `y` na primeira, `n` na segunda, `y` na terceira.
- **Resultado Esperado:** Apenas a primeira e terceira ocorrências são alteradas.
- **Covers:** Confirmação interativa gc.

---

## Casos de Teste de Performance

### TC18 — Varredura Instantânea em Buffer Extenso (< 5 ms em 50.000 linhas)
- **Alvo:** `internal/search/engine.go`
- **Tipo:** performance
- **Resultado Esperado:** Varredura concluída em menos de 5 ms.

### TC19 — Latência de Digitação Incremental a 60 FPS (< 16 ms por frame)
- **Alvo:** `internal/app/model.go`
- **Tipo:** performance
- **Resultado Esperado:** Tempo de processamento e renderização por caractere digitado inferior a 16 ms.
