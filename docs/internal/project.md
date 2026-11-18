# Projeto Balde: Gestor Financeiro Pessoal CLI

## Sumário Executivo

A ideia é desenvolver um gestor financeiro pessoal em linha de comando (CLI), utilizando Go ou Rust, que armazene os dados em SQLite ou arquivos de texto/Markdown. O projeto visa ser particularmente "amigável a agentes de IA" como o Opencode, permitindo importação fácil de extratos bancários e de cartões de crédito. A metodologia de gestão financeira será baseada rigorosamente na "Teoria dos Buckets", com referências em:
- [Budget with Buckets — Everything Guide](https://www.budgetwithbuckets.com/guide/everything/)
- [Budget Buckets App — Method Guide](https://budgetbucket.app/guide/budget-buckets-method)
- [Spaceship — How to Bucket Your Money](https://www.spaceship.com.au/learn/how-to-bucket-your-money/)

**Análise de Oportunidade:**

*   **Existência de Ferramentas Similares:** Existem CLIs de finanças pessoais em Go (ex: `jms-guy/greed`, `isaaczidany/personal-finance-cli`, `newtoallofthis123/moni`), e algumas mencionam "savings buckets". No entanto, nenhuma foca explicitamente em "AI-friendly" para agentes ou na adesão estrita à metodologia `budgetwithbuckets.com`.
*   **Oportunidade Open-Source:** Há uma clara oportunidade para um projeto open-source que se diferencie pela sua "AI-friendliness", importação flexível de dados e implementação precisa da Teoria dos Buckets.

## Plano de Desenvolvimento Detalhado

### 1. Definição Detalhada da Estrutura de Dados

*   **Buckets:** Nome, limite, saldo, categoria, regras associadas.
*   **Transações:** Data, descrição, valor, categoria, conta de origem/destino, ID único.
*   **Contas:** Nome, tipo (corrente, poupança, crédito), saldo atual, limite de crédito (para cartões).
*   **Armazenamento:** SQLite é a opção recomendada para robustez, capacidade de consulta e escalabilidade. Arquivos MD/TXT podem ser utilizados para relatórios e exportação.

### 2. Design da Interface CLI

*   **Comandos Essenciais:**
    *   `balde init`: Inicializa o projeto e configura o armazenamento.
    *   `balde config`: Gerencia configurações do aplicativo.
    *   `balde account add <nome> <tipo> <saldo_inicial>`: Adiciona uma nova conta.
    *   `balde bucket add <nome> <limite> <categoria>`: Cria um novo bucket.
    *   `balde transaction add <valor> <descricao> <conta> <bucket>`: Adiciona uma transação manual.
    *   `balde transaction import <arquivo> --format <csv|ofx|qif>`: Importa transações de arquivos.
    *   `balde transaction categorize <id> <categoria>|<bucket>`: Categoriza ou associa uma transação a um bucket.
    *   `balde allocate <valor> <bucket>`: Aloca fundos para um bucket.
    *   `balde rain`: Mostra o "rain" disponível (dinheiro a orçar) e permite alocar rapidamente entre os buckets ("make it rain").
    *   `balde report <tipo> [--bucket <nome>] [--category <nome>] [--period <mes|ano>]`: Gera relatórios.
    *   `balde view buckets`: Exibe o estado atual dos buckets.
    *   `balde view transactions`: Exibe as transações.
*   **Output Amigável para IA:** Implementar opções para saída em JSON ou CSV para comandos de listagem e relatório, facilitando o parse por agentes.
*   **Input Amigável para IA:** Comandos claros e parametrizados, permitindo que agentes construam facilmente interações.

### 3. Módulo de Importação de Dados

*   **Suporte a Formatos:** Desenvolver parsers para CSV, OFX e QIF.
*   **Mapeamento de Campos:** Um mecanismo para o usuário ou agente mapear colunas/campos dos arquivos importados para os campos de transação internos.
*   **Categorização Assistida por IA:** Um sistema inicial de regras configuráveis para sugerir categorias e buckets com base na descrição da transação. Agentes podem interagir para treinar ou ajustar essas regras.

### 4. Lógica de Buckets (baseada na metodologia de envelope/bucket budgeting)

Referências:
- [Budget with Buckets — Everything Guide](https://www.budgetwithbuckets.com/guide/everything/)
- [Budget Buckets App — Method Guide](https://budgetbucket.app/guide/budget-buckets-method)
- [Spaceship — How to Bucket Your Money](https://www.spaceship.com.au/learn/how-to-bucket-your-money/)

Princípios:
*   **Buckets como containers de propósito:** 5-8 buckets que agrupam despesas relacionadas (ex: "Moradia & Contas", "Alimentação & Transporte"). Não usar categorias granulares.
*   **Frequência Global Única:** Suportar ciclos semanais, quinzenais ou mensais. Todos os valores convertidos para a frequência escolhida.
*   **Alocação de Fundos ("Make it Rain"):** Gerenciamento da alocação de receita em buckets via "Rain" (dinheiro a orçar). A metáfora é um barril de chuva capturando toda a receita que é então distribuída para os buckets menores. No início de cada mês o usuário "faz chover" — distribui cada centavo do rain entre seus buckets.
*   **Controle de Saldo:** Rastreamento do saldo disponível em cada bucket.
*   **Transferências:** Permitir transferências entre buckets.
*   **Regras de Gastos:** Encorajar gastos do bucket apropriado e alertar sobre despesas que excedam o limite do bucket.
*   **Regra 50/30/20 como guia:** Sugerir alocações iniciais (50% necessidades, 30% desejos, 20% poupança/dívidas).
*   **Simplicidade por design:** Desencorajar mais de 8 buckets para evitar abandono.
*   **Flexibilidade de estrutura:** Suportar tanto setups minimalistas (ex: Barefoot Investor — 3 buckets: Blow/Gasto, Mojo/Emergência, Grow/Crescimento) quanto setups mais detalhados (ex: 8 buckets com contas variáveis, fixas, viagens, investimentos, presentes, etc.). O sistema deve ser adaptável ao perfil do usuário.
*   **Automação de alocação:** Suportar transferências automáticas percentuais ou fixas para cada bucket ao receber receita, simulando o "set and forget" de múltiplas contas bancárias.
*   **Rebalanceamento:** Permitir ajustes periódicos nas alocações quando um bucket consistentemente estoura ou sobra — os valores são "números vivos" que evoluem.

### 5. Relatórios e Visualização

*   **Relatórios Básicos:** Desempenho dos buckets, gastos por categoria, fluxo de caixa.
*   **Exportação:** Opções de exportação para CSV, JSON e Markdown estruturado para facilitar a análise por agentes ou outras ferramentas.

### 6. Escolha da Linguagem e Estrutura do Projeto

*   **Linguagem:** **Go** — escolhida pela simplicidade, velocidade de compilação, binário estático único e excelente ecossistema para CLIs.
*   **Bibliotecas Go:**
    *   **CLI:** `cobra` para parsing de argumentos e estrutura de comandos.
    *   **SQLite:** `modernc.org/sqlite` (pure Go, sem CGO) ou `mattn/go-sqlite3`.
    *   **Data Parsing:** `encoding/csv` (stdlib), bibliotecas para OFX/QIF.
    *   **JSON:** `encoding/json` (stdlib).
*   **Organização do Projeto:** Módulos bem definidos para `core` (lógica de negócio), `cli` (interface), `data` (armazenamento), `import` (parsers), `report` (geração de relatórios).
