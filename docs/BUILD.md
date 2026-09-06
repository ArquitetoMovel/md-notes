# Infraestrutura de Build Multiplataforma — md-notes (mdn)

Este documento descreve detalhadamente a infraestrutura de compilação, distribuição e automação do **md-notes** (`mdn`) para as três principais plataformas: **macOS**, **Linux** e **Windows**.

---

## 1. Visão Geral das Plataformas e Arquiteturas

O `mdn` é compilado sem dependências de CGO (`CGO_ENABLED=0`), gerando binários estáticos, portáteis e sem vínculos dinâmicos com libc/glibc, garantindo funcionamento em qualquer distribuição Linux (incluindo Alpine com musl), versões modernas do macOS e do Windows.

| Plataforma | Arquitetura | Nome do Artefato | Formato de Pacote |
|---|---|---|---|
| **macOS (Darwin)** | Apple Silicon (`arm64`) | `bin/darwin_arm64/mdn` | `.tar.gz` |
| **macOS (Darwin)** | Intel (`amd64`) | `bin/darwin_amd64/mdn` | `.tar.gz` |
| **macOS (Darwin)** | Universal (`arm64` + `x86_64`) | `bin/darwin_universal/mdn` | `.tar.gz` |
| **Linux** | 64-bit x86 (`amd64`) | `bin/linux_amd64/mdn` | `.tar.gz` |
| **Linux** | 64-bit ARM (`arm64`) | `bin/linux_arm64/mdn` | `.tar.gz` |
| **Windows** | 64-bit x86 (`amd64`) | `bin/windows_amd64/mdn.exe` | `.zip` |
| **Windows** | 64-bit ARM (`arm64`) | `bin/windows_arm64/mdn.exe` | `.zip` |

---

## 2. Otimizações de Compilação

Para atender aos requisitos de alta performance e tamanho reduzido (< 25 MB do PRD; binários finais ficam entre **3.3 MB e 3.7 MB**):

- `CGO_ENABLED=0`: Binário 100% estático em puro Go.
- `-trimpath`: Remove referências a caminhos de arquivos absolutos da máquina do desenvolvedor (reprodutibilidade e rastreabilidade limpa).
- `-ldflags="-s -w"`:
  - `-s`: Remove informações de símbolos de depuração DWARF.
  - `-w`: Desativa geração de tabelas de símbolos DWARF.
  - Redução de ~35% no tamanho final do executável.
- **Injeção de Metadados em Tempo de Build**:
  - `main.Version`: Versão semântica ou tag git (`git describe`).
  - `main.GitCommit`: Hash SHA resumido do git commit.
  - `main.BuildDate`: Timestamp ISO 8601 UTC.

Executando `mdn --version` no terminal:
```bash
$ mdn --version
mdn version v1.0.0 (78d517c, 2026-09-03T01:52:53Z, darwin/arm64)
```

---

## 3. Formas de Compilação

A infraestrutura fornece **3 interfaces complementares** para atender a qualquer desenvolvedor ou ambiente CI/CD:

### Método A: Usando `make` (Linux / macOS / WSL)

O `Makefile` é a interface padrão recomendada para desenvolvedores Unix:

```bash
# Ver todos os alvos disponíveis com descrições
make help

# Compilar para a plataforma e arquitetura atual da máquina
make build

# Compilar todas as 6 combinações de SO e arquiteturas
make build-all

# Gerar o Universal Binary para macOS (Apple Silicon + Intel via lipo)
make universal-darwin

# Compilar tudo, empacotar (.tar.gz / .zip) e gerar checksums SHA256 em dist/
make package

# Rodar a suíte de testes unitários
make test

# Rodar testes com geração de relatório HTML de cobertura
make test-coverage

# Limpar artefatos de compilação (bin/ e dist/)
make clean
```

---

### Método B: Usando Go Builder Puro (`scripts/build.go`)

Ideal para **Windows nativo** (onde `make` ou `bash` podem não estar instalados) ou qualquer sistema com apenas o compilador Go instalado:

```bash
# Compilar todas as plataformas padrão
go run scripts/build.go

# Compilar apenas para Windows x64
go run scripts/build.go -os=windows -arch=amd64

# Compilar para Linux ARM64
go run scripts/build.go -os=linux -arch=arm64

# Compilar e empacotar arquivos finais com hash SHA256 (usa compressão nativa Go)
go run scripts/build.go -package

# Definir versão customizada e limpar diretórios
go run scripts/build.go -version=1.2.0 -clean -package
```

---

### Método C: Script Shell Avançado (`scripts/build.sh`)

Script POSIX/Bash com suporte a parâmetros e saída colorida:

```bash
# Compilar todas as plataformas
./scripts/build.sh

# Compilar plataforma específica
./scripts/build.sh --os linux --arch amd64

# Compilar e empacotar para release
./scripts/build.sh -v 1.0.0 --package

# Ver opções
./scripts/build.sh --help
```

---

## 4. Estrutura dos Artefatos de Distribuição (`dist/`)

Após executar `make package` ou `go run scripts/build.go -package`, o diretório `dist/` conterá:

```
dist/
├── checksums.txt
├── mdn-v1.0.0-darwin-amd64.tar.gz
├── mdn-v1.0.0-darwin-arm64.tar.gz
├── mdn-v1.0.0-darwin-universal.tar.gz
├── mdn-v1.0.0-linux-amd64.tar.gz
├── mdn-v1.0.0-linux-arm64.tar.gz
├── mdn-v1.0.0-windows-amd64.zip
└── mdn-v1.0.0-windows-arm64.zip
```

O arquivo `checksums.txt` permite aos usuários e scripts de instalação verificar a integridade criptográfica de cada download:

```bash
# Verificação no Linux
sha256sum -c checksums.txt

# Verificação no macOS
shasum -a 256 -c checksums.txt
```

---

## 5. Integração Contínua (CI/CD)

Foram configurados dois fluxos no GitHub Actions em `.github/workflows/`:

1. **`ci.yml`**:
   - Disparado em todo `push` ou `pull_request` nas branches principais (`main`, `master`).
   - Matriz de testes nativos executados simultaneamente em máquinas virtuais **Ubuntu**, **macOS** e **Windows**.
   - Validação da compilação cruzada completa e geração de artefatos de teste.

2. **`release.yml`** & **`.goreleaser.yaml`**:
   - Disparado automaticamente na publicação de tags de versão (ex: `git tag v1.0.0 && git push --tags`).
   - Utiliza o [GoReleaser](https://goreleaser.com) para compilar, assinar/gerar checksums, criar release notes automáticos a partir dos commits e disponibilizar os arquivos para download direto na aba **Releases** do GitHub.
