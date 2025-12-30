# Backend Challenge

[![Tests](https://github.com/v-venes/backend-challenge/actions/workflows/tests.yml/badge.svg)](https://github.com/v-venes/backend-challenge/actions/workflows/tests.yml)
[![codecov](https://codecov.io/gh/v-venes/backend-challenge/branch/main/graph/badge.svg)](https://codecov.io/gh/v-venes/backend-challenge)
[![License](https://img.shields.io/github/license/v-venes/backend-challenge)](LICENSE)

**Table of Contents**

- [Backend Challenge](#backend-challenge)
  - [Descrição](#descrição)
  - [Fluxo de Ingestão](#fluxo-de-ingestão)
  - [Banco de Dados](#banco-de-dados)
    - [Schema do banco](#schema-do-banco)
        - [trade_at e trade_date](#trade_at-e-trade_date)
        - [Índices](#índices)
  - [Como rodar o projeto](#como-rodar-o-projeto)
    - [Pré requisitos](#pré-requisitos)
    - [Rodando o projeto](#rodando-o-projeto)
  - [Pontos para melhoria](#pontos-para-melhoria)

## Descrição

Resolução do desafio de backend, esse projeto é constituido por duas partes ingestão / leitura, optei pela linguagem Go pois poderia usar as goroutines para salvar no banco de forma concorrente, assim eu conseguiria ir lendo e formatando as linhas do CSV, jogando para algum agregador que iria ir enviando um grupo de stocks para as goroutines. Para persistir os dados eu optei pelo PostgreSQL visando a consistência dos dados, os índices para as consultas agregadas e para idempotência, um índice é utilizado para validar a unicidade do registro (ticker + data e hora fechamento) quando ocorre um conflito eu utilizo o DO NOTHING, assim é possível rodar a ingestão várias vezes sem duplicar um registro, além disso o segundo índice é utilizado para otimizar as consultas. 

## Fluxo de Ingestão

O Fluxo de ingestão acontece da seguinte forma:

1. 📥 É disparada um goroutine para servir de batcher, ele utiliza dois canais, um para receber as stocks formatadas e outra para enviar uma quantidade de stocks para ser inserida no banco através dos workers.

2. ⚙️ Os workers são disparados para gravar os batches de stocks no banco.
> A quantidade de workers é definida da seguinte forma `runtime.NumCPU() * 2`

3. 📂 A aplicação começa a percorrer todos os arquivos zip dentro da pasta `./data`.
> É possível obter os arquivos na área de [Cotações da B3](https://www.b3.com.br/pt_br/market-data-e-indices/servicos-de-dados/market-data/cotacoes/cotacoes/).

4. 📄 O CSV é lido linha a linha, para cada linha a aplicação transforma o array de string na struct que representa o stock e após isso envia para o canal que o batcher fica escutando.

5. 🔄 Sempre que um novo stock é enviado pelo canal, o batcher adicionar em um array e verifica se já é possível enviar o array para ser persistido no banco.

6. 💾 Quando o batcher envia o array de stocks, um dos workers pega o array e segue para persistir os dados no banco.

## Banco de Dados

### Schema do banco

| id (PK)   | ticker      | trade_at   | trade_date   | price         | quantity   |  
|-----------|-------------|------------|--------------|---------------|------------| 
| bigserial | varchar(16) | timestampz | date         | decimal(15,3) | int8       | 

#### trade_at e trade_date

A príncipio eu estava utilizando só a coluna trade_at, que possui a DataReferencia+HoraFechamento, porém nas consultas eu estava precisando calcular o DATE() toda hora e isso estava prejudicando a performance da consulta, adicionando a coluna de trade_date eu precisei ter um índice separado para ela e a coluna ticker, fazendo o include de price e quantity que são utilizados na agregação.

#### Índices

* **idx_stocks_ticker_trade_at**: Índice utilizado para validar unicidade do registro, quando ocorre conflito ele não adiciona nada novo.
* **idx_stocks_ticker_trade_date_inc**: Índice covering utilizado para as agregações nas consultas, inclui as colunas price e quantity.

## Como rodar o projeto

### Pré requisitos

1. Você precisa ter o [Docker](https://docs.docker.com/engine/install/) e o [Docker Compose](https://docs.docker.com/compose/install) instalados

2. Também é necessário ter o comando `make` instalado no sistema operacional
> A maioria das distros Linux disponibiliza o comando através do pacote `build-essential`

3. Adicionar envs no projeto

Crie um arquivo `.env` na raiz do projeto e adicionar o seguinte conteúdo:

```env
POSTGRES_HOST=db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=stocks
POSTGRES_PORT=5432
GO_ENV=production
```

### Rodando o projeto

1. 🚀 Subir dependências
```bash
make up
```

2. 🗃️ Rodar migrations
```bash
make migrate
```

3. 📦 Adicionar arquivos para ingestão

Copie os arquivos ZIP para: `./data`

> Você pode baixar os arquivos na área de [Cotações da B3](https://www.b3.com.br/pt_br/market-data-e-indices/servicos-de-dados/market-data/cotacoes/cotacoes/).

4. ⚡ Rodar ingestão
```bash
make ingest
```

5. 🌐 Subir API
```bash
make api
```
> Endpoint de consulta disponível em: http://localhost:8080/stocks/aggregate

6. 📊 Exemplos de consultas

```bash
curl --request GET \
  --url 'http://localhost:8080/stocks/aggregate?ticker=WINZ25&data_inicio=2025-12-01'
```

ou sem o parâmetro `data_inicio`:

```bash
curl --request GET \
  --url 'http://localhost:8080/stocks/aggregate?ticker=PETR4'
```

## Pontos para melhoria
Aqui são alguns pontos que foram surgindo enquanto eu desenvolvia e testava

### 1. 🚀 Processamento paralelo de arquivos
Atualmente estou lendo um arquivo por vez, o consumo de memória está baixo e o tempo de ingestão não está tão alto, mas fiquei pensando se processamento paralelamente os arquivos o tempo de ingestão diminuiria, mas também precisaria ficar atento ao uso de memória.

### 2. 🔄 Procurar outra forma de garantir idempotência
Pensei no índice `idx_stocks_ticker_trade_at` para verificar os conflitos, porém isso prejudica a escrita.  

### 3. 💾 Pensar em outra forma de salvar os dados para não precisar de duas colunas com o "mesmo valor"
Vem de encontro com o ponto acima, eu utilizo a coluna `trade_at` para validar se um registro ja existe, mas isso me atrapalho na consulta, pois ou eu adicionava uma coluna nova só com a data (o que de fato ocorreu) ou eu precisaria usar o DATE() nas consultas que prejudicaria a performance.