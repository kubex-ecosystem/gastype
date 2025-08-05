# 🌟 **Pipeline GasType Bitwise**

## “Legível no debug, insano em produção.”

---

## **1️⃣ ETAPA 1 — Transpilação SEM OFUSCAÇÃO**

**Objetivo:** Converter para a lógica bitwise otimizada, mas ainda **com nomes e estrutura legível** para humanos.

**Ações no código:**

* Booleans → bitflags (`uint64` + masks).
* `if/else chains` → jump tables.
* `switch` → lookup tables.
* Literais sensíveis → placeholders legíveis (`const Secret = "<protected:admin>"`).
* **Mantém nomes originais** de funções, structs, variáveis.

**CLI:**

```bash
gastype transpile --mode optimize --no-obfuscate -i ./src -o ./out_optimized
```

**Saída esperada:**

* Código otimizado **legível**.
* Estrutura igual ao original, mas com lógica transpilada.
* Diretório `./out_optimized/`.

---

## **2️⃣ ETAPA 2 — Validação & Teste**

**Objetivo:** Garantir que o código otimizado funciona exatamente como o original.

**Ações:**

* Executar testes unitários e integração no **código otimizado legível**.
* Comparar resultados com baseline (antes da transpilação).
* Gerar relatório de compatibilidade.

**CLI:**

```bash
gastype validate --baseline ./src --optimized ./out_optimized --tests ./tests
```

**Saída esperada:**

* Relatório `validation_report.json` com:

  ```json
  {
    "total_tests": 120,
    "passed": 118,
    "failed": 2,
    "coverage": "98.3%"
  }
  ```

* Lista de arquivos/trechos aprovados para próxima etapa.

---

## **3️⃣ ETAPA 3 — Ofuscação Seletiva**

**Objetivo:** Aplicar **máxima ofuscação** apenas onde **já foi validado**.

**Controle granular com comentários no código:**

```go
//gastype:nobfuscate
func DebugHandler() { ... }

//gastype:force
func UltraSecureAuth() { ... }
```

**Ações no código:**

* Renomear funções, variáveis e tipos (`UltraSecureAuth` → `_X91z_f1`).
* Converter strings sensíveis para byte arrays.
* Reordenar blocos internos (flatten control flow).
* Opcional: gerar **dispatch tables** para funções.

**CLI:**

```bash
gastype obfuscate --from ./out_optimized --only-passed --marks
```

**Saída esperada:**

* Código **otimizado + ofuscado** apenas onde permitido.
* Diretório `./out_obfuscated/`.

---

## **4️⃣ ETAPA 4 — Build Final Otimizado**

**Objetivo:** Produzir binário final **insano de rápido e difícil de reverter**.

**Ações:**

* Compilar com otimizações de Go (`-gcflags=all="-B" -trimpath`).
* Strip de debug symbols (`-ldflags="-s -w"`).
* Opcional: compressão UPX.
* Gerar checksum e artefatos para distribuição.

**CLI:**

```bash
gastype build --source ./out_obfuscated --final --compress
```

**Saída esperada:**

* Binário final `./dist/myapp`.
* Checksums SHA256.
* Relatório `build_report.json` com:

  ```json
  {
    "binary_size": "8.9MB",
    "startup_time": "18ms",
    "memory_usage": "28MB",
    "throughput_gain": "38%"
  }
  ```

---

## **🚀 Integração no CI/CD**

YAML para rodar no GitHub Actions:

```yaml
jobs:
  gastype-pipeline:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Transpile (no obfuscation)
        run: gastype transpile --mode optimize --no-obfuscate -i ./src -o ./out_optimized

      - name: Validate optimized code
        run: gastype validate --baseline ./src --optimized ./out_optimized --tests ./tests

      - name: Obfuscate passed components
        run: gastype obfuscate --from ./out_optimized --only-passed

      - name: Build final binary
        run: gastype build --source ./out_obfuscated --final --compress
```

---

## **📊 Benefícios do Pipeline**

* **Debug seguro** → sempre existe versão legível otimizada.
* **Rollback rápido** → só remove a etapa de ofuscação se der ruim.
* **Controle absoluto** → decide o que fica legível ou não.
* **Ganhos reais** → performance + redução de footprint + segurança.
