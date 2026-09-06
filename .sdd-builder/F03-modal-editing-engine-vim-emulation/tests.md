# Especificação de Casos de Teste: F03. Modal Editing Engine (Vim Emulation)

## Sumário de Cobertura

| Tipo de Teste | Quantidade | Foco / Alvo |
|---|---|---|
| Unitário | 10 | Movimentos de cursor, contadores, histórico Undo/Redo, registradores e comandos Ex |
| Integração | 7 | Máquina de estados, transições de modo, clipboard do SO, comandos compostos e integrações |
| E2E / TUI | 3 | Sessões de edição modal completas no modelo Bubble Tea |

---

## Matriz de Rastreabilidade

| Critério de Aceitação (PRD Seção 9) | Casos de Teste |
|---|---|
| O editor inicia no modo NORMAL com cursor em bloco | TC01 |
| As teclas `h`, `j`, `k`, `l`, `w`, `b`, `0`, `$`, `gg`, `G` movimentam o cursor com precisão pelas coordenadas do texto | TC02, TC03, TC04 |
| As teclas `i`, `a`, `I`, `A`, `o`, `O` entram no modo INSERT na posição e linha corretas | TC05 |
| A tecla `Esc` retorna do modo INSERT ou VISUAL para o modo NORMAL | TC06 |
| O comando `x` apaga o caractere sob o cursor e `dd` apaga a linha inteira | TC07, TC08 |
| O comando `yy` copia a linha atual e `p` cola o conteúdo após a linha/cursor | TC09, TC10 |
| O texto copiado com `yy` ou `y` (em modo Visual) é copiado para a área de transferência do sistema operacional | TC11 |
| A tecla `u` desfaz a última alteração e `Ctrl+r` refaz a alteração revertida | TC12, TC13 |
| A tecla `:` posiciona o cursor na linha de comando inferior para digitação de comandos ex | TC14, TC15 |
| Error Handling: comando ex inválido (`E492`) e limites de Undo/Redo | TC16, TC17 |
| Full Scope: Contadores numéricos (`5j`, `3dd`), buscas `f/F/t/T` e modo visual em bloco | TC18, TC19, TC20 |

---

## Casos de Teste Unitários

### TC01 — Inicialização no Modo Normal
- **Alvo:** `internal/vim/engine.go` (`NewEngine`)
- **Tipo:** unit
- **Pré-condições:** Buffer instanciado.
- **Passos:**
  1. Instanciar `NewEngine(buf)`.
  2. Verificar `engine.State.CurrentMode`.
- **Resultado Esperado:** O modo inicial é `ModeNormal` com cursor posicionado em `(0, 0)`.
- **Cobre:** Critério PRD: editor inicia no modo NORMAL com cursor em bloco.

### TC02 — Movimentação Direcional Básica (`h`, `j`, `k`, `l`)
- **Alvo:** `internal/vim/motion.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com 3 linhas de texto em UTF-8: `"Linha 1 🚀"`, `"Linha 2"`, `"Linha 3"`.
- **Passos:**
  1. Pressionar `l` até o final da linha 1.
  2. Pressionar `j` para descer para a linha 2.
  3. Pressionar `h` para recuar colunas.
  4. Pressionar `k` para subir para a linha 1.
- **Resultado Esperado:** O cursor se desloca célula a célula sobre runes Unicode sem ultrapassar os limites superior, inferior, esquerdo e direito do texto.
- **Cobre:** Critério PRD: navegação direcional precisa por coordenadas.

### TC03 — Navegação por Palavras (`w` e `b`)
- **Alvo:** `internal/vim/motion.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"const [count, setCount] = useState(0);"`
- **Passos:**
  1. Pressionar `w` sucessivamente a partir do início.
  2. Pressionar `b` sucessivamente a partir do final.
- **Resultado Esperado:** `w` salta para o início da próxima palavra ou símbolo (`[`, `count`, `,`, `setCount`, `]`, `=`, `useState`, `(`, `0`, `)`, `;`). `b` retorna exatamente para os inícios das palavras anteriores.
- **Cobre:** Critério PRD: movimentação por palavras com `w` e `b`.

### TC04 — Limites de Linha e Documento (`0`, `$`, `gg`, `G`)
- **Alvo:** `internal/vim/motion.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com 10 linhas. Cursor na linha 5, coluna 4.
- **Passos:**
  1. Pressionar `0` e verificar coluna.
  2. Pressionar `$` e verificar coluna.
  3. Pressionar `gg` e verificar linha e coluna.
  4. Pressionar `G` e verificar linha.
- **Resultado Esperado:** `0` move para col 0; `$` move para o último caractere da linha; `gg` move para linha 0, col 0; `G` move para a última linha do buffer.
- **Cobre:** Critério PRD: movimentação por limites de linha e arquivo.

### TC05 — Operadores de Entrada no Modo Insert (`i`, `a`, `I`, `A`, `o`, `O`)
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"Texto"` com cursor na coluna 2 (`'x'`).
- **Passos:**
  1. Testar `i`: entra no modo Insert na coluna 2.
  2. Testar `a`: entra no modo Insert na coluna 3.
  3. Testar `I`: entra no modo Insert na coluna 0.
  4. Testar `A`: entra no modo Insert na coluna 5 (fim da linha).
  5. Testar `o`: cria nova linha abaixo e posiciona o cursor nela.
  6. Testar `O`: cria nova linha acima e posiciona o cursor nela.
- **Resultado Esperado:** Todas as teclas alternam para `ModeInsert` posicionando o cursor com exatidão conforme a semântica do Vim.
- **Cobre:** Critério PRD: comandos de entrada no modo INSERT na posição e linha corretas.

### TC06 — Retorno para o Modo Normal via `Esc`
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** Engine em `ModeInsert` ou `ModeVisualChar`.
- **Passos:**
  1. Pressionar `Esc`.
- **Resultado Esperado:** O modo ativo retorna para `ModeNormal`, qualquer seleção visual é limpa e o cursor ajusta a coluna caso esteja além do final da linha.
- **Cobre:** Critério PRD: tecla Esc retorna do modo INSERT ou VISUAL para NORMAL.

### TC07 — Operação de Deleção de Caractere (`x`)
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** Linha `"Teste"` com cursor na coluna 1 (`'e'`).
- **Passos:**
  1. Pressionar `x`.
- **Resultado Esperado:** O caractere `'e'` é removido, a linha passa a ser `"Tste"` e o caractere apagado é salvo no registrador.
- **Cobre:** Critério PRD: comando x apaga caractere sob o cursor.

### TC08 — Operação de Deleção de Linha (`dd`) e Palavra (`dw`)
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com 3 linhas: `"Primeira"`, `"Segunda"`, `"Terceira"`.
- **Passos:**
  1. Posicionar na linha 1 e pressionar `dd`.
  2. Posicionar na linha restante e pressionar `dw`.
- **Resultado Esperado:** `dd` remove a linha inteira salvando com flag linewise no registrador; `dw` remove a palavra até o próximo limite.
- **Cobre:** Critério PRD: comando dd apaga linha inteira e dw apaga palavra.

### TC09 — Operação de Cópia (`yy`, `yw`) e Colagem (`p`, `P`)
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com texto `"Linha copiada"`.
- **Passos:**
  1. Pressionar `yy` na linha.
  2. Pressionar `p` na linha.
  3. Pressionar `P` na linha.
- **Resultado Esperado:** `yy` copia a linha para o registrador; `p` cola como uma nova linha abaixo; `P` cola como uma nova linha acima.
- **Cobre:** Critério PRD: comando yy copia linha e p cola conteúdo.

### TC10 — Pilha de Histórico de Desfazer (`u`) e Refazer (`Ctrl+r`)
- **Alvo:** `internal/vim/history.go`
- **Tipo:** unit
- **Pré-condições:** Buffer com texto original `"Inicial"`.
- **Passos:**
  1. Digitar texto em modo Insert: `"Inicial + Modificado"`.
  2. Pressionar `Esc` (registra snapshot).
  3. Pressionar `u`.
  4. Pressionar `Ctrl+r`.
- **Resultado Esperado:** `u` reverte o buffer para `"Inicial"`; `Ctrl+r` restaura o buffer para `"Inicial + Modificado"`.
- **Cobre:** Critério PRD: tecla u desfaz e Ctrl+r refaz alterações.

---

## Casos de Teste de Integração

### TC11 — Sincronização com o Clipboard do Sistema Operacional
- **Alvo:** `internal/vim/clipboard.go` & `internal/vim/engine.go`
- **Tipo:** integration
- **Pré-condições:** Engine instanciada com clipboard integrado.
- **Passos:**
  1. Executar `yy` em uma linha contendo `"Texto do Clipboard"`.
  2. Ler o conteúdo do clipboard via `clipboard.ReadAll()`.
- **Resultado Esperado:** O conteúdo `"Texto do Clipboard\n"` está presente no clipboard do sistema operacional.
- **Cobre:** Critério PRD: texto copiado com yy ou y é sincronizado com a área de transferência do SO.

### TC12 — Desfazer Atômico de Sessão Contínua de Inserção
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** integration
- **Pré-condições:** Modo Normal em buffer vazio.
- **Passos:**
  1. Pressionar `i` e digitar `"Primeira Linha\nSegunda Linha\nTerceira Linha"`.
  2. Pressionar `Esc`.
  3. Pressionar `u`.
- **Resultado Esperado:** O buffer retorna integralmente ao estado vazio anterior à entrada no modo Insert com um único `u`.
- **Cobre:** Agrupamento atômico de sessões no modo Insert.

### TC13 — Descarte Circular de Histórico no Limite de 200 Snapshots
- **Alvo:** `internal/vim/history.go`
- **Tipo:** integration
- **Pré-condições:** Histórico inicializado.
- **Passos:**
  1. Executar 250 mutações consecutivas com snapshots.
  2. Inspecionar o tamanho da pilha de histórico.
- **Resultado Esperado:** O histórico retém exatamente os 200 estados mais recentes sem estourar memória.
- **Cobre:** Capacidade PRD: limite de 200 checkpoints de undo/redo por buffer.

### TC14 — Entrada no Modo Command (`:`) e Digitação de Comando
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** integration
- **Pré-condições:** Modo Normal ativo.
- **Passos:**
  1. Pressionar `:`.
  2. Digitar `w notas.md`.
- **Resultado Esperado:** O modo passa para `ModeCommand`, `State.CommandInput == "w notas.md"` e o cursor se posiciona na barra inferior.
- **Cobre:** Critério PRD: tecla : abre linha de comando inferior.

### TC15 — Execução de Comandos Ex `:w`, `:q`, `:wq` e `:q!`
- **Alvo:** `internal/vim/command.go`
- **Tipo:** integration
- **Pré-condições:** Engine conectada a um buffer.
- **Passos:**
  1. Executar `:w` e verificar comando de salvamento disparado.
  2. Executar `:q` e verificar comando de saída disparado.
  3. Executar `:wq` e verificar sequência de salvar e sair.
  4. Executar `:q!` e verificar saída forçada com flag `Force = true`.
- **Resultado Esperado:** Todas as diretivas Ex são interpretadas e mapeadas para as mensagens correspondentes (`SaveFileMsg`, `QuitMsg`).
- **Cobre:** Critério PRD: execução de comandos ex :w, :q, :wq, :q!.

### TC16 — Tratamento de Comando Ex Inválido (`E492`)
- **Alvo:** `internal/vim/command.go`
- **Tipo:** integration
- **Pré-condições:** Modo Command ativo.
- **Passos:**
  1. Digitar `:comando_invalido` e pressionar `Enter`.
- **Resultado Esperado:** Retorna ao modo Normal e define a mensagem na barra de status: `E492: Não é um comando de editor: comando_invalido`.
- **Cobre:** Error Handling PRD: comando ex desconhecido.

### TC17 — Notificação de Limites da Pilha de Undo / Redo
- **Alvo:** `internal/vim/history.go` & `internal/vim/engine.go`
- **Tipo:** integration
- **Pré-condições:** Buffer sem edições anteriores.
- **Passos:**
  1. Pressionar `u`.
  2. Pressionar `Ctrl+r`.
- **Resultado Esperado:** Ao pressionar `u`, exibe `Já na alteração mais antiga`; ao pressionar `Ctrl+r`, exibe `Já na alteração mais recente`.
- **Cobre:** Error Handling PRD: fim da pilha de undo/redo.

---

## Casos de Teste Full Scope

### TC18 — Multiplicadores Numéricos de Repetição (`5j`, `3dd`, `10w`)
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** unit / integration
- **Pré-condições:** Buffer com 20 linhas.
- **Passos:**
  1. Digitar `5` e depois `j`.
  2. Digitar `3` e depois `dd`.
- **Resultado Esperado:** O cursor avança exatamente 5 linhas para baixo; `3dd` apaga 3 linhas consecutivas em uma única operação atômica.
- **Cobre:** Full Scope PRD: suporte a contadores numéricos de repetição.

### TC19 — Busca de Caractere na Linha Ativa (`f`, `F`, `t`, `T`)
- **Alvo:** `internal/vim/motion.go`
- **Tipo:** unit
- **Pré-condições:** Linha contendo `"chave = 'valor_configurado'"`. Cursor na coluna 0.
- **Passos:**
  1. Pressionar `f` e depois `'`.
  2. Pressionar `t` e depois `_`.
  3. Pressionar `F` e depois `=`.
- **Resultado Esperado:** `f'` salta diretamente para a primeira aspas simples; `t_` salta para o caractere imediatamente antes do sublinhado; `F=` salta para trás até o sinal de igual.
- **Cobre:** Full Scope PRD: comandos de busca de caractere na linha f/F/t/T.

### TC20 — Modos Visuais por Caractere (`v`), Linha (`V`) e Bloco (`Ctrl+v`)
- **Alvo:** `internal/vim/engine.go`
- **Tipo:** integration
- **Pré-condições:** Buffer de texto com 4 linhas.
- **Passos:**
  1. Pressionar `v` e mover o cursor (seleção de caracteres).
  2. Pressionar `V` e mover o cursor (seleção de linhas completas).
  3. Pressionar `Ctrl+v` e mover o cursor (seleção em bloco retangular).
  4. Pressionar `y` em cada modo.
- **Resultado Esperado:** Os limites da seleção correspondem precisamente ao modo visual ativo e a cópia extrai o formato de texto correto para o registrador.
- **Cobre:** Full Scope PRD: seleção visual em bloco e por linhas.
