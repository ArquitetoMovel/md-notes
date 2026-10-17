# Casos de Teste: F09. Seleção com mouse e transferência de texto

## Resumo de cobertura

| Tipo | Casos | Alvo |
|---|---:|---|
| Unitário | 9 | Regras de coordenadas, seleção, conteúdo, limites, buffer e histórico. |
| Integração | 8 | Contratos consumidos de F01, F03 e F07; roteamento TEA e falhas de clipboard. |
| Ponta a ponta | 3 | Fluxos visíveis de selecionar/copiar, colar texto externo e substituir/desfazer. |

Os testes automatizados usam um provedor falso de clipboard. O comportamento de `Cmd` interceptado pelo emulador e a interação com `tmux` exigem ensaio manual no terminal, pois esses eventos não chegam ao processo durante testes unitários.

## Rastreabilidade

| Critério do PRD, seção 9 | Casos |
|---|---|
| A1 — Arrastar seleciona texto bruto em várias linhas com gutter, rolagem e `word_wrap` | TC01, TC02, TC10, TC18 |
| A2 — Wrap visual não acrescenta `\n`; quebra lógica permanece | TC03, TC11 |
| A3 — `Cmd+C`/`Ctrl+C` encaminhado copia UTF-8 e marcadores Markdown | TC03, TC11, TC18 |
| A4 — `Ctrl+C` sem seleção conserva saída protegida | TC12 |
| A5 — Paste externo em INSERT é literal | TC08, TC17, TC19 |
| A6 — Paste sobre seleção substitui só o intervalo, marca dirty e desfaz em uma etapa | TC05, TC06, TC17, TC20 |
| A7 — Copiar não marca dirty; `Esc` cancela sem editar | TC09, TC11, TC18 |
| A8 — Até 10.000 caracteres ou 200 linhas preservam conteúdo e quebras | TC04, TC05, TC19 |
| A9 — Clipboard vazio/indisponível e excesso não causam inserção parcial | TC07, TC13, TC14, TC15, TC16 |
| I1 — Dados de F01 definem conteúdo selecionado e ponto de inserção | TC05, TC10, TC17 |
| I2 — Modo, seleção, histórico e clipboard de F03 habilitam cópia, substituição e undo | TC06, TC11, TC17, TC20 |
| I3 — Mapeamento de F07 funciona com gutter, rolagem e wrap | TC01, TC02, TC10 |
| A10 — Clique direito com seleção copia sem editar; falha preserva seleção | TC21, TC23 |
| A11 — Clique direito sem seleção cola no ponto clicado e desfaz em uma etapa | TC22, TC23 |

## Extensão: botão direito

### TC21 — Clique direito com seleção
- **Alvo:** `TestApp_RightClickClipboard` em `internal/app/model_vim_test.go`.
- **Resultado esperado:** Uma única escrita do Markdown bruto no clipboard, seleção limpa após sucesso, buffer e cursor inalterados; falha de escrita preserva a seleção.

### TC22 — Colagem contextual em área vazia
- **Alvos:** `TestViewport_MouseToPasteCursor`, `TestEngine_PasteClipboardAt` e `TestApp_RightClickClipboard`.
- **Resultado esperado:** Clique após o texto ou abaixo da última linha cola na posição lógica correta em NORMAL/INSERT; uma operação `u` restaura o documento anterior; barra de status e modos COMMAND/SEARCH não colam.

### TC23 — Falhas da colagem contextual
- **Alvo:** `TestApp_RightClickClipboard` em `internal/app/model_vim_test.go`.
- **Resultado esperado:** Clipboard vazio, leitura indisponível e payload acima do limite exibem a mensagem pertinente sem alterar texto, cursor, histórico ou seleção.

## Casos unitários

### TC01 — Coordenadas de mouse e limites da área de texto
- **Alvo:** `TestViewport_MouseToCursor` em `internal/ui/viewport_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Viewport com gutter, scrollbar, linhas visíveis e `TopLine` diferente de zero.
- **Passos:** Executar subcasos para clique no início, no gutter, após o último caractere, na margem direita e fora das linhas exibidas.
- **Resultado esperado:** Cada célula mapeia para uma posição lógica válida; gutter e espaço após o texto são limitados às extremidades; statusline não vira seleção.
- **Cobre:** A1, I3.

### TC02 — Unicode largo e linha quebrada visualmente
- **Alvo:** `TestViewport_MouseWideRuneAndWrap` em `internal/ui/viewport_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Linha com acentos, emoji e caractere de largura 2 atravessando uma quebra visual.
- **Passos:** Arrastar em ambas as direções pelos segmentos e nas duas células do caractere largo.
- **Resultado esperado:** Nenhuma posição interna de rune é criada; `StartCol`/`EndCol`, cursor e seleção coincidem com as células renderizadas.
- **Cobre:** A1, I3.

### TC03 — Texto copiado é o Markdown bruto
- **Alvo:** `TestEngine_CopySelection` em `internal/vim/engine_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Buffer com `#`, `**`, `|`, UTF-8, duas linhas lógicas e uma quebra apenas visual; clipboard falso.
- **Passos:** Selecionar intervalo que atravessa as linhas e copiar.
- **Resultado esperado:** Clipboard recebe os marcadores e caracteres originais, exatamente uma quebra lógica e nenhuma quebra adicional de wrap.
- **Cobre:** A2, A3.

### TC04 — Limites inclusivos de tamanho
- **Alvo:** `TestEngine_PasteModesAndLimits` em `internal/vim/engine_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Buffer limpo e payloads de texto UTF-8 gerados em memória.
- **Passos:** Testar exatamente 10.000 runes, 10.001 runes, 200 linhas e 201 linhas em subcasos independentes.
- **Resultado esperado:** Os limites exatos são aceitos sem perda; valores acima do limite são recusados antes de modificar buffer ou histórico.
- **Cobre:** A8, A9.

### TC05 — Substituição de intervalo e terminações de linha
- **Alvo:** `TestBuffer_ReplaceRangeAndLineEndings` em `internal/buffer/buffer_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Buffer com prefixo e sufixo não selecionados, linhas CRLF e cursor dentro da seleção.
- **Passos:** Inserir texto externo multilinha sobre um intervalo e repetir para colagem sem seleção.
- **Resultado esperado:** Apenas o intervalo muda; prefixo, sufixo e terminações de linhas preexistentes permanecem; novas linhas adotam a convenção definida; cursor vai ao fim da inserção e dirty fica verdadeiro.
- **Cobre:** A6, A8, I1.

### TC06 — Um undo remove somente a colagem
- **Alvo:** `TestEngine_PasteAtomicUndo` em `internal/vim/engine_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Histórico inicializado, digitação anterior ainda pendente em INSERT e seleção ativa em outro subcaso.
- **Passos:** Colar texto multilinha, acionar `u` uma vez e depois `Ctrl+r`.
- **Resultado esperado:** Um undo reverte somente a colagem e preserva a digitação anterior; redo restaura toda a colagem em uma etapa, sem checkpoint por rune.
- **Cobre:** A6, I2.

### TC07 — Erros estritos do provedor de clipboard
- **Alvo:** `TestClipboard_StrictReadWriteErrors` em `internal/vim/clipboard_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Provedor falso configurável para erro de leitura ou escrita.
- **Passos:** Solicitar cópia e colagem nos dois modos de falha; consultar registrador e histórico.
- **Resultado esperado:** Erros são retornados ao chamador; não há confirmação falsa de cópia nem alteração do buffer; o fallback legado permanece disponível a `yy`/`p`.
- **Cobre:** A9.

### TC08 — Paste literal e modos permitidos
- **Alvo:** `TestEngine_PasteModesAndLimits` em `internal/vim/engine_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Payload contendo `Esc`, `:q`, `dd`, `#` e quebras de linha; modos INSERT, NORMAL, COMMAND e SEARCH.
- **Passos:** Enviar evento de paste nos modos permitidos e bloqueados; enviar `Ctrl+V` em NORMAL sem seleção.
- **Resultado esperado:** Texto é inserido literalmente em INSERT ou sobre seleção; modos bloqueados não alteram documento; `Ctrl+V` sem seleção conserva Visual Block.
- **Cobre:** A5.

### TC09 — Seleção e cópia não modificam o documento
- **Alvo:** `TestApp_MouseSelectionLifecycle` em `internal/app/model_vim_test.go`.
- **Tipo:** unitário.
- **Pré-condições:** Buffer limpo com seleção iniciada por mouse.
- **Passos:** Arrastar, soltar, copiar, iniciar nova seleção e cancelar com `Esc`; repetir cancelamento por clique simples fora dela.
- **Resultado esperado:** `IsDirty` e texto não mudam; destaque persiste após soltar e some após copiar ou cancelar; cursor permanece válido.
- **Cobre:** A7.

## Casos de integração

### TC10 — Mouse, viewport e buffer
- **Alvo:** `TestApp_MouseSelectionLifecycle` cruzando F07 → F01 em `internal/app/model_vim_test.go`.
- **Tipo:** integração.
- **Pré-condições:** Modelo com gutter, linhas suficientes para rolagem e `word_wrap` ativo.
- **Passos:** Enviar `tea.MouseMsg` Press, Motion e Release sobre linhas visuais distintas, inclusive após rolar.
- **Resultado esperado:** Seleção F03 aponta para o intervalo exato de runes F01 correspondente às células F07; roda continua rolando sem criar seleção.
- **Cobre:** A1, I1, I3.

### TC11 — Seleção Vim e área de transferência
- **Alvo:** `TestApp_CopyShortcutAndQuitFallback` cruzando F03 → clipboard em `internal/app/model_vim_test.go`.
- **Tipo:** integração.
- **Pré-condições:** Seleção visual ativa, buffer limpo e clipboard falso.
- **Passos:** Enviar evento `Ctrl+C` ou equivalente reconhecido pelo editor.
- **Resultado esperado:** Texto bruto, UTF-8 e quebra lógica chegam ao clipboard uma vez; não ocorre saída, alteração do buffer ou checkpoint de undo.
- **Cobre:** A2, A3, A7, I2.

### TC12 — `Ctrl+C` sem seleção mantém saída protegida
- **Alvo:** `TestApp_CopyShortcutAndQuitFallback` em `internal/app/model_vim_test.go`.
- **Tipo:** integração.
- **Pré-condições:** Nenhuma seleção; subcasos com buffer limpo e dirty.
- **Passos:** Enviar `Ctrl+C`.
- **Resultado esperado:** Buffer limpo encerra como antes; buffer dirty mostra `E37` e permanece aberto.
- **Cobre:** A4.

### TC13 — Clipboard vazio
- **Alvo:** `TestApp_PasteRouting` em `internal/app/model_vim_test.go`.
- **Tipo:** integração negativa.
- **Pré-condições:** INSERT ou seleção ativa, clipboard falso retorna texto vazio.
- **Passos:** Enviar atalho explícito de colar.
- **Resultado esperado:** Exibe `Área de transferência sem texto para colar`; buffer, seleção, dirty e histórico permanecem iguais.
- **Cobre:** A9, erro de clipboard vazio do PRD.

### TC14 — Erro de leitura do clipboard
- **Alvo:** `TestApp_PasteRouting` em `internal/app/model_vim_test.go`.
- **Tipo:** integração negativa.
- **Pré-condições:** Provedor falso retorna erro de leitura.
- **Passos:** Enviar atalho explícito de colar.
- **Resultado esperado:** Exibe `Não foi possível ler a área de transferência`; nada é inserido e não há checkpoint novo.
- **Cobre:** A9, erro de leitura do PRD.

### TC15 — Erro de escrita do clipboard
- **Alvo:** `TestApp_CopyShortcutAndQuitFallback` em `internal/app/model_vim_test.go`.
- **Tipo:** integração negativa.
- **Pré-condições:** Seleção ativa e provedor falso retorna erro de escrita.
- **Passos:** Enviar `Ctrl+C`.
- **Resultado esperado:** Exibe `Não foi possível copiar para a área de transferência`; seleção e cursor permanecem; não sai, não marca dirty e não apresenta sucesso falso.
- **Cobre:** A9, erro de cópia do PRD.

### TC16 — Payload acima do limite sem inserção parcial
- **Alvo:** `TestApp_PasteRouting` em `internal/app/model_vim_test.go`.
- **Tipo:** integração negativa.
- **Pré-condições:** Seleção ativa e payload de 10.001 runes ou 201 linhas.
- **Passos:** Enviar `KeyMsg.Paste` e, em subcaso, atalho explícito de colar.
- **Resultado esperado:** Exibe `Texto para colar excede 10.000 caracteres ou 200 linhas`; seleção, buffer, dirty e histórico permanecem intocados.
- **Cobre:** A9, erro de limite do PRD.

### TC17 — Paste delimitado atravessa os contratos F01/F03/F07
- **Alvo:** `TestApp_PasteRouting` e `TestApp_SettingsBlocksTransfer` em `internal/app/model_vim_test.go`.
- **Tipo:** integração.
- **Pré-condições:** Modelo em INSERT ou com seleção ativa, `tea.KeyMsg` de paste com Markdown multilinha; subcasos com modal aberto, COMMAND e SEARCH.
- **Passos:** Enviar um evento delimitado, verificar seleção, buffer, viewport e histórico; repetir nos modos bloqueados.
- **Resultado esperado:** Um único evento insere/substitui literalmente, atualiza viewport e dirty, e cria um passo de undo; painel e prompts não deixam o payload atingir o documento.
- **Cobre:** A5, A6, I1, I2.

## Casos ponta a ponta

### TC18 — Selecionar e copiar com mouse
- **Alvo:** Sessão TUI com terminal que encaminha mouse e atalho de cópia.
- **Tipo:** ponta a ponta.
- **Pré-condições:** Arquivo Markdown aberto sem alterações, clipboard observável.
- **Passos:** Arrastar sobre título e parágrafo, soltar, acionar `Cmd+C` mapeado ou `Ctrl+C`, e conferir conteúdo do clipboard.
- **Resultado esperado:** Destaque segue o arraste; clipboard contém Markdown bruto correto; o arquivo continua limpo e o editor não encerra.
- **Cobre:** A1, A3, A7.

### TC19 — Colar texto vindo de outro programa
- **Alvo:** Sessão TUI e área de transferência do sistema.
- **Tipo:** ponta a ponta.
- **Pré-condições:** Texto UTF-8 com marcadores Markdown e quebras de linha copiado em outro programa, dentro dos limites do PRD.
- **Passos:** Entrar em INSERT, colar pelo atalho configurado no terminal, observar buffer e salvar/reabrir o arquivo.
- **Resultado esperado:** Texto é inserido literalmente; marcadores e quebras lógicas persistem; o indicador `[+]` aparece antes de salvar.
- **Cobre:** A5, A8.

### TC20 — Substituir seleção e desfazer
- **Alvo:** Sessão TUI com histórico de edição.
- **Tipo:** ponta a ponta.
- **Pré-condições:** Arquivo com texto anterior e trecho selecionado pelo mouse.
- **Passos:** Colar texto externo, acionar `u` uma vez e `Ctrl+r` uma vez.
- **Resultado esperado:** Apenas o trecho selecionado é substituído; `u` restaura o documento anterior inteiro e `Ctrl+r` reaplica a substituição inteira.
- **Cobre:** A6, I2.

## Fora do escopo automatizado

- Verificar em macOS, Linux e Windows quais terminais entregam ou consomem `Cmd+C`, `Cmd+V` e `Ctrl+V`; quando consumidos, configurar o mapeamento do terminal e repetir TC18/TC19.
- Repetir o arraste e a colagem real dentro de `tmux` com mouse habilitado, porque o multiplexador pode capturar eventos antes do Bubble Tea.
