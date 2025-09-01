# ![Gastype Banner](assets/top_banner_ct_md.png)

---

**Uma ferramenta de transpilação e análise de código baseada em Go AST, projetada para otimizar performance, reduzir o tamanho do binário e aumentar a segurança do código através de transformações bitwise e ofuscação.**

---

## **Índice**

1. [Sobre o Projeto](#1-sobre-o-projeto)
2. [Recursos](#2-recursos)
3. [Instalação](#3-instalação)
4. [Uso da CLI](#4-uso-da-cli)
      - [Verificação de Tipos](#verificação-de-tipos)
      - [Transpilação](#transpilação)
      - [Pipeline de Transpilação Staged](#pipeline-de-transpilação-staged)
5. [Transpilação e Otimização](#5-transpilação-e-otimização)
6. [Contribuindo](#6-contribuindo)
7. [Licença](#7-licença)
8. [Contato](#8-contato)

---

### **1. Sobre o Projeto**

Gastype (parte do projeto Kubex) é uma ferramenta de linha de comando poderosa e flexível para engenheiros de software Go. Ela se destaca na análise estática de código fonte (AST) para identificar oportunidades de otimização e transformar código automaticamente. O objetivo principal é melhorar a performance, segurança e eficiência de binários Go, especialmente para ambientes de produção.

A ferramenta oferece um conjunto de "passes" de transpilação que convertem padrões de código comuns e menos eficientes em operações bitwise ultra-otimizadas.

### **2. Recursos**

- **Verificação de Tipos Paralela**: Executa verificações de tipo em múltiplos arquivos Go simultaneamente para garantir a qualidade do código em projetos grandes.
- **Análise de Otimização**: Analisa código e sugere otimizações, como converter campos `bool` em `structs` para `flags` bitwise.
- **Transpilação Automática**: Transforma automaticamente o código fonte, aplicando passes de otimização para gerar um binário de alta performance.
- **Ofuscação de Código**: Ofusca nomes de variáveis, nomes de funções e literais de string para aumentar a segurança e dificultar a engenharia reversa.
- **Otimização de Estruturas de Controle**: Converte longas cadeias de `if/else` em "tabelas de salto" para execução mais rápida.
- **Pipeline Staged**: Fornece um fluxo de trabalho multi-estágio (`staged-transpilation`) que inclui transpilação, validação, ofuscação e build final.

### **3. Instalação**

**Requisitos**: Go versão 1.19 ou posterior.

```bash
# Clone o repositório
git clone https://github.com/rafa-mori/gastype.git
cd gastype

# Build e instale o binário
make install
```

### **4. Uso da CLI**

A CLI `gastype` é o ponto central para todas as operações.

#### **Verificação de Tipos**

- **`gastype check`**: Inicia a verificação de tipos em arquivos Go em um diretório específico.
- **`gastype watch`**: Monitora um diretório por mudanças de arquivos e automaticamente dispara a verificação de tipos.

**Exemplos**:

```bash
# Executa verificação de tipos no diretório atual com 4 workers
gastype check -d ./ -w 4 -o results.json

# Monitora um projeto e envia notificações por email em caso de erro
gastype watch --dir ./my-project --email user@example.com --notify
```

#### **Transpilação**

- **`gastype transpile`**: O comando principal para transformar código. Suporta vários modos de operação.

**Exemplos**:

```bash
# Analisa um projeto e exibe oportunidades de otimização
gastype transpile --input ./src --mode analyze --format text

# Transpila um único arquivo, convertendo campos bool para flags bitwise
gastype transpile --input ./config.go --output ./config_optimized.go --mode transpile --passes bool-to-flags

# Transpila um projeto inteiro com máxima ofuscação
gastype transpile --input ./my-app --output ./my-app-optimized --mode full-project --security 3
```

#### **Pipeline de Transpilação Staged**

`gastype` fornece um pipeline de otimização de quatro estágios para garantir a robustez do código.

1. **`gastype transpile --no-obfuscate`** (Estágio 1: Transpilação Limpa)
      - Cria uma versão otimizada do código sem ofuscação, ideal para debug e testes.
2. **`gastype validate`** (Estágio 2: Validação)
      - Garante que o código otimizado se comporta de forma idêntica ao original executando testes e verificando o build.
3. **`gastype obfuscate`** (Estágio 3: Ofuscação Seletiva)
      - Aplica ofuscação apenas às partes do código que passaram na validação.
4. **`gastype build`** (Estágio 4: Build Final)
      - Compila o binário final, aplicando otimizações do compilador Go, removendo símbolos de debug e opcionalmente comprimindo o binário com UPX.

### **5. Transpilação e Otimização**

O coração do `gastype` está em seus "passes" de transpilação. Cada passe é uma transformação AST modular que pode ser habilitada independentemente.

**Exemplos de Otimização**:

- **`bool-to-flags`**: Converte structs com múltiplos campos `bool` em um único campo `uint64` com flags bitwise, reduzindo o consumo de memória e melhorando a localidade de cache.
- **`jump-table`**: Transforma declarações `if/else` encadeadas que comparam a mesma variável em um mapa de funções, resultando em execução mais rápida.
- **`string-obfuscate`**: Substitui literais de string por arrays de bytes, tornando a análise estática do binário mais difícil.

### **6. Contribuindo**

Apreciamos seu interesse em contribuir para o `gastype`. Sinta-se à vontade para abrir `issues` ou submeter `pull requests`. Por favor, consulte o [Guia de Contribuição](https://github.com/rafa-mori/gastype/blob/main/CONTRIBUTING.md) para mais detalhes.

### **7. Licença**

Este projeto está licenciado sob a Licença MIT.

### **8. Contato**

- **Desenvolvedor**: Rafael Mori ([faelmori@gmail.com](mailto:faelmori@gmail.com))
- **GitHub**: [https://github.com/rafa-mori](https://github.com/rafa-mori)
- **GitHub (2)**: [https://github.com/rafa-mori](https://github.com/rafa-mori)
