# Cadastro de Assistidos

Sistema das filiais para acompanhar Assistidos, Núcleos Familiares e Atendimentos, com painel de indicadores.

## Pessoas

**Assistido**:
A pessoa acompanhada pela organização, criança ou adulta, com ou sem oficina, com ou sem cesta.
_Avoid_: beneficiário, aluno (como tipo da pessoa), criança (como tipo), cliente

**Aluno**:
Assistido com pelo menos uma Oficina vigente.
_Avoid_: usar Aluno para quem só tem cesta ou só Atendimento

**Responsável Legal**:
Nome que identifica o adulto de referência do Núcleo quando há criança. Não é Assistido só por constar na ficha.
_Avoid_: pai, mãe, tutor (como entidades distintas)

**Operador**:
Pessoa autenticada que registra fichas, consulta e lê o painel. Não é Assistido.
_Avoid_: usuário, admin, conta (como sinônimo da pessoa acompanhada)

## Agrupamento

**Núcleo Familiar**:
Agrupamento real de um ou mais Assistidos que compartilham casa. Independente é Núcleo com um Assistido. Cesta Básica, endereço e WhatsApp pertencem ao Núcleo.
_Avoid_: família (como texto da checkbox), grupo, household

**Unidade**:
Filial à qual o Assistido está vinculado. O Núcleo não tem Unidade.
_Avoid_: sede, loja, campus

**Oficina**:
Atividade vigente na ficha do Assistido (matrícula atual, não histórico do dia).
_Avoid_: curso, turma, vínculo genérico, workshop

**Cesta Básica**:
Flag vigente no Núcleo: a família está ou não na lista agora. Não é o registro de uma entrega.
_Avoid_: elegível vs recebe como dois status, benefício, kit

## Registro

**Cadastro**:
A ficha vigente do Assistido (identidade interna). Volta à mesma pessoa é pela consulta, não por fusão automática. CPF, se preenchido, é único.
_Avoid_: ficha solta sem Assistido, match por nome

**Atendimento**:
Ocorrência datada ligada a um Assistido (relato, itens entregues, data). Cada salvamento da ficha atualiza o Cadastro e acrescenta um Atendimento.
_Avoid_: visita, evento, ocorrência (como nome canônico)
