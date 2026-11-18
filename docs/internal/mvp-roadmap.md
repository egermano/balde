# Projeto Balde: MVP e Roadmap de Desenvolvimento

## Produto Mínimo Viável (MVP)

O objetivo do MVP é entregar uma ferramenta CLI funcional e útil que demonstre o conceito central da gestão por buckets e a "AI-friendliness", permitindo o gerenciamento básico de finanças pessoais.

### Funcionalidades do MVP:

1.  **Estrutura de Dados Essencial:**
    *   Implementação do modelo de dados para **Contas**, **Buckets** e **Transações** usando **Go** e **SQLite**.
2.  **Internacionalização (i18n):**
    *   O MVP será em **inglês** (interface, mensagens, labels de buckets sugeridos).
    *   A arquitetura deve suportar i18n desde o início usando um sistema de dicionários/locales (ex: `go-i18n`).
    *   Strings da interface, mensagens de erro, labels de buckets sugeridos e textos de help devem ser externalizados em arquivos de locale (ex: `locales/en.json`, `locales/pt-BR.json`).
    *   Dicionários adicionais (pt-BR, es, etc.) serão adicionados em fases posteriores via roadmap.
    *   O comando `balde config set locale <code>` permite trocar o idioma da interface.
3.  **Configuração de Moeda e Formatação Numérica:**
    *   O usuário deve poder configurar o separador decimal (`,` ou `.`) e o separador de milhar durante `balde init` ou via `balde config`.
    *   Exemplo: R$ 1.234,56 (BR) vs $1,234.56 (US).
    *   O símbolo da moeda deve ser configurável (R$, $, €, £, etc.).
    *   Internamente os valores continuam armazenados como inteiros (centavos). A formatação acontece apenas na camada de apresentação (`cli` e `report`).
    *   Configurações armazenadas no arquivo de config do projeto (locale + currency symbol + decimal separator + thousands separator).
4.  **Interface CLI Básica:**
    *   `balde init`: Inicializa o banco de dados SQLite e a estrutura do projeto.
    *   `balde account add <nome> <tipo> <saldo_inicial>`: Permite adicionar novas contas bancárias/de crédito.
    *   `balde bucket add <nome> <limite>`: Cria novos buckets com um limite definido.
    *   `balde transaction add <valor> <descricao> <conta> <bucket>`: Adiciona transações manualmente, vinculando-as a uma conta e um bucket.
    *   `balde allocate <valor> <bucket>`: Aloca manualmente uma quantia para um bucket (e.g., após receber o salário).
    *   `balde rain`: Mostra o "rain" disponível (dinheiro a orçar) e inicia o processo de "make it rain" — alocação rápida entre os buckets.
    *   `balde view buckets`: Exibe o nome do bucket, limite e saldo atual.
    *   `balde view transactions`: Lista todas as transações, incluindo conta e bucket associado.
3.  **Importação Simples de Dados:**
    *   `balde transaction import <arquivo> --format csv`: Suporte a importação de transações via arquivo CSV com um mapeamento de colunas básico e pré-definido.
4.  **Lógica Essencial de Buckets:**

    Baseada na metodologia de envelope/bucket budgeting, com referências em:
    - [Budget with Buckets — Everything Guide](https://www.budgetwithbuckets.com/guide/everything/)
    - [Budget Buckets App — Method Guide](https://budgetbucket.app/guide/budget-buckets-method)
    - [Spaceship — How to Bucket Your Money](https://www.spaceship.com.au/learn/how-to-bucket-your-money/)

    Princípios centrais do método:

    *   **Buckets como containers de propósito:** O dinheiro é organizado por propósito (não por categorias granulares). Em vez de dezenas de subcategorias, usa-se 5-8 buckets que agrupam despesas relacionadas (ex: "Moradia & Contas", "Alimentação & Transporte", "Fundo de Emergência").
    *   **Frequência Global:** O sistema deve suportar uma frequência única de alocação (semanal, quinzenal ou mensal) que corresponda ao ciclo de receita do usuário. Todos os valores de bucket são expressos nessa frequência. Contas mensais em setups quinzenais devem ser divididas (mensal / 2.17), etc.
    *   **Dinheiro a Orçar (Rain / To Be Budgeted - TBB):** O sistema deve calcular o total de dinheiro disponível que ainda não foi alocado a nenhum bucket. A metáfora é um barril de chuva (rain barrel) que captura toda a receita — o "rain" é o saldo total das contas menos os saldos já alocados nos buckets. É o dinheiro disponível para "fazer chover" sobre os buckets.
    *   **"Make it Rain" (Distribuição Mensal):** No início de cada mês, o usuário distribui o "rain" entre seus buckets, enchendo cada um até sua "fill line" (valor alvo). Este é o momento central do método — é quando o usuário decide para onde vai cada centavo. O sistema deve facilitar esse processo com comandos como `balde rain` que mostra o rain disponível e permite alocar rapidamente.
    *   **Alocação de Fundos:** Funcionalidade para mover dinheiro do TBB para os buckets, definindo o valor disponível para cada grupo de gastos/poupança.
    *   **Gastos a Partir dos Buckets:** Toda transação de saída deve ser vinculada a um bucket, e o valor deduzido do saldo desse bucket. O dinheiro flui naturalmente dentro do bucket — gastar mais no supermercado não importa, desde que esteja dentro do bucket "Alimentação & Transporte".
    *   **Simplicidade por design:** Máximo de 5-8 buckets. O sistema deve desencorajar a criação de buckets demais, pois isso é o principal motivo de abandono segundo o método.
    *   **Flexibilidade de estrutura:** Suportar tanto setups minimalistas (ex: Barefoot Investor — Blow/Mojo/Grow) quanto setups detalhados (8+ buckets). Adaptável ao perfil do usuário.
    *   **Automação de alocação (básica):** Suportar alocação percentual ou fixa por bucket ao receber receita, simulando "auto-transfer" entre contas.
    *   **Rebalanceamento:** Permitir ajustes nas alocações quando um bucket consistentemente estoura ou sobra — os valores são "números vivos".
    *   **Regra 50/30/20 como guia:** O sistema pode sugerir alocações iniciais baseadas na regra (50% necessidades, 30% desejos, 20% poupança/dívidas), mas o usuário tem liberdade total para personalizar.
    *   **Buckets Essenciais sugeridos para começar:**
        - Moradia & Contas (rent/mortgage, utilities)
        - Alimentação & Transporte (groceries, fuel, transport)
        - Seguros & Fixed Bills (insurance, phone)
        - Pessoal & Lazer (dining out, hobbies, streaming)
        - Fundo de Emergência (build to 3-6 months)
    *   **Visualização de Saldos:** Exibição clara do saldo atual de cada bucket, mostrando quanto ainda está disponível para gastar.
    *   **Tratamento de Saldo Negativo (Básico):** Indicação visual ou alerta se um bucket estiver com saldo negativo (overspent). A resolução (e.g., cobrir de outro bucket) pode ser implementada em fases posteriores, mas o reconhecimento é MVP.
5.  **Output Amigável para IA:**
    *   Todos os comandos `view` (e potencialmente outros) devem ter uma opção `--json` para produzir saída formatada em JSON, facilitando o consumo por agentes de IA.
6.  **Design para Integração com AI Harnesses:**
    *   Implementar a CLI com uma arquitetura modular que facilite a extensão e integração com ferramentas externas ou agentes de IA.
    *   Garantir que os comandos e suas saídas (especialmente em JSON) sejam padronizados e facilmente parseáveis, permitindo que agentes de IA construam e executem comandos de forma programática.
    *   Considerar a criação de um "modo agente" no CLI que possa oferecer prompts específicos ou formatos de saída otimizados para interação com IA.

## Roadmap de Desenvolvimento

### Fase 1: Expansão e Usabilidade (Pós-MVP)

*   **Novos Dicionários de i18n:**
    *   Adicionar locales para pt-BR, es, fr, de, etc.
    *   Criar sistema de contribuição de dicionários pela comunidade.
    *   Traduzir labels de buckets sugeridos, mensagens de erro e textos de help.
*   **Sistema de Taxonomia (Tags/Categorias):**
    *   Adicionar uma camada de categorização independente dos buckets e contas. Uma transação pertence a um bucket e uma conta, mas pode receber múltiplas tags de uma taxonomia hierárquica.
    *   Exemplo: Bucket "Transporte", tags `#carro` ou `#metro`. Bucket "Alimentação", tags `#supermercado`, `#restaurante`, `#delivery`.
    *   Suportar taxonomia hierárquica: `#transporte.carro`, `#transporte.metro`, `#alimentacao.supermercado`.
    *   Comandos: `balde tag add <nome>`, `balde tag list`, `balde transaction tag <id> <tag1> <tag2>...`.
    *   Permitir filtros e relatórios por tag: `balde report by-tag transporte`, `balde view transactions --tag carro`.
    *   Isso permite granularidade na análise sem comprometer a simplicidade dos buckets — os buckets continuam 5-8, mas as tags permitem drill-down ilimitado.
*   **Múltiplos Formatos de Importação:**
    *   Adicionar suporte para importação de arquivos OFX e QIF.
    *   Melhorar a importação CSV com mapeamento de colunas configurável (via CLI ou arquivo de configuração).
*   **Categorização Assistida por Regras:**
    *   Implementar um sistema de regras configuráveis (e.g., "se descrição contém 'Starbucks', categorizar como 'Café'").
    *   Comando `balde transaction categorize` para aplicar regras e ajustar manualmente.
*   **Relatórios Básicos:**
    *   `balde report summary`: Resumo do estado financeiro geral.
    *   `balde report by-bucket`: Detalhes de gastos por bucket.
    *   `balde report by-category`: Detalhes de gastos por categoria (se implementado).
*   **Exportação de Dados:**
    *   Exportar dados de transações e buckets para CSV e Markdown.
*   **Melhorias na UX do CLI:**
    *   Mensagens de erro mais claras.
    *   Completamento automático de comandos (se a biblioteca CLI suportar).

### Fase 2: Inteligência e Automação

*   **Categorização Inteligente:**
    *   Desenvolver um módulo de aprendizado de máquina simples para sugerir categorias e/ou buckets com base em transações anteriores e descrições.
    *   Feedback do usuário para melhorar o modelo (e.g., `balde transaction correct-category <id> <nova_categoria>`).
*   **Transações Recorrentes:**
    *   Configuração e gerenciamento de transações que ocorrem regularmente (salário, aluguel, assinaturas).
*   **Metas Financeiras:**
    *   Definição de metas para buckets (e.g., poupança para viagem) e acompanhamento do progresso.
*   **Controle de Dívidas:**
    *   Funcionalidade para rastrear dívidas (empréstimos, dívidas de cartão de crédito) e planos de pagamento.
*   **Visualização Aprimorada:**
    *   Geração de gráficos simples no terminal ou exportação de dados para ferramentas de visualização externas.

### Fase 3: Ecossistema e Integrações

*   **Integração com APIs Bancárias:**
    *   Explorar integrações com APIs (como Plaid ou APIs de Open Banking) para importação automática de extratos (com as devidas considerações de segurança e privacidade).
*   **Sistema de Plugins:**
    *   Permitir que a comunidade crie e compartilhe plugins para novos formatos de importação, relatórios personalizados ou outras funcionalidades.
*   **UI Adicional:**
    *   Considerar a criação de uma interface de usuário web simples ou um wrapper para desktop, para aqueles que preferem uma experiência mais visual, mantendo o CLI como o core.
*   **Multi-usuário/Compartilhamento:**
    *   (Se aplicável) Implementar funcionalidades para gerenciar finanças de múltiplas pessoas ou compartilhar com parceiros/família.

Este roadmap visa fornecer uma direção clara para o desenvolvimento do Projeto Balde, começando com um MVP sólido e expandindo para um conjunto abrangente de funcionalidades.
