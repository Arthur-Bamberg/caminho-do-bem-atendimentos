# PRD: PWA de Cadastro de Assistidos

## Problem Statement

A equipe das filiais registra crianças e adultos, oficinas, cesta básica e o que aconteceu no dia numa tela única, mas o protótipo atual não autentica, mistura pessoa/casa/visita no mesmo envio, usa fundo bege em tudo e não deixa gravar ficha incompleta com clareza. Sem um Cadastro estável, a volta do Assistido não liga o Atendimento de novembro ao de setembro, irmãos viram duas “famílias”, e o painel (famílias, assistidos, cestas, oficinas) não reflete o que a operação precisa contar.

## Solution

Uma PWA responsiva em React, autenticada, com API em Go, fundos brancos e roxo/bege só em detalhes. O Operador entra, preenche a ficha (todos os campos opcionais), grava Assistido + Núcleo Familiar + Atendimento, reencontra a pessoa na consulta (identidade interna; CPF único se existir) e lê o painel de indicadores filtrado por Unidade. Irmãos entram no mesmo núcleo por escolha explícita. Cesta, endereço e WhatsApp são do núcleo. Oficina é matrícula vigente do Assistido.

## User Stories

1. As an Operador, I want to sign in with a provisioned account, so that only staff can see fichas and indicadores.
2. As an Operador, I want the app to reject unauthenticated access to cadastro, consulta and painel, so that data does not leak on a shared phone.
3. As an Operador, I want to sign out, so that the next person on the device does not inherit my session.
4. As an Operador, I want a password manager–friendly login (visible labels, paste allowed), so that I am not blocked by cognitive or autofill-hostile auth.
5. As an Operador, I want to install the app on the home screen (PWA), so that I can open cadastro quickly on a phone in the filial.
6. As an Operador, I want layouts that work on a phone and a desktop, so that I can register in the field and review the painel in the office.
7. As an Operador, I want a cadastro form with the same fields as the attached first screen, so that the daily routine does not change.
8. As an Operador, I want every form field to be optional, so that I can save a ficha even when the family has no CPF, phone, or full address.
9. As an Operador, I want visible labels (not placeholder-only), so that I still know what each field is after I start typing.
10. As an Operador, I want to save a Cadastro with no name and see it listed as “Sem nome”, so that incomplete arrivals are not blocked.
11. As an Operador, I want CPF to remain optional, so that children and adults without document can still be Assistidos.
12. As an Operador, I want the system to reject a second Assistido with the same CPF when a CPF is present, so that I do not silently duplicate a person.
13. As an Operador, I want each new save from a blank form to create a new Assistido (internal identity), so that two homonymous siblings without CPF stay as two people.
14. As an Operador, I want to pick an existing Núcleo Familiar on the ficha, so that a second sibling joins the same house instead of creating another núcleo.
15. As an Operador, I want to leave núcleo search and responsável empty, so that an independent adult becomes a núcleo of one.
16. As an Operador, I want Responsável Legal to stay a display name on the núcleo, so that the adult who signs is not automatically an Assistido.
17. As an Operador, I want Unidade to belong to the Assistido, so that “Assistidos por unidade” counts people, not houses.
18. As an Operador, I want siblings in the same núcleo to be allowed different Unidades, so that one child at Centro and one at Norte still share cesta and address.
19. As an Operador, I want address and WhatsApp to live on the Núcleo, so that updating the mother’s phone updates both siblings.
20. As an Operador, I want the second sibling, once linked to a núcleo, to inherit address and WhatsApp, so that I do not retype the house.
21. As an Operador, I want to mark zero or more Oficinas on the Assistido, so that current enrolment is on the ficha.
22. As an Operador, I want changing Oficinas on a later save to replace the current set, so that the painel “Alunos por oficina” is always the now.
23. As an Operador, I want an Assistido with at least one Oficina to count as Aluno, so that the oficina chart does not include people with no enrolment.
24. As an Operador, I want a single Cesta Básica flag on the Núcleo (“on the list now”), so that two siblings do not count as two cestas.
25. As an Operador, I want unchecking cesta to drop the núcleo from the cesta KPI immediately, so that the painel matches the current list, not delivery history.
26. As an Operador, I want each save of a ficha to update the Cadastro and append an Atendimento, so that November’s visit does not erase September’s relato.
27. As an Operador, I want Atendimento date to default to today and remain optional to edit, so that I am not forced to pick a date under pressure.
28. As an Operador, I want Relato da Situação / Itens Entregues stored on the Atendimento, so that deliveries of the day are history, not a field that overwrites the house.
29. As an Operador, I want to open Consulta, search by nome, CPF or responsável, and filter by Unidade, so that I can find the same Assistido on return.
30. As an Operador, I want to see zero results clearly when nothing matches, so that I know to create a new Cadastro instead of editing the wrong person.
31. As an Operador, I want pagination on consulta, so that a large filial list remains usable on a phone.
32. As an Operador, I want to open an existing Assistido from consulta into the ficha, so that the next save updates that Cadastro and adds a new Atendimento.
33. As an Operador, I want the painel to show Famílias impactadas as count of Núcleos that have at least one Assistido in the selected Unidade filter (or all Unidades), so that two siblings count as one família.
34. As an Operador, I want Assistidos totais to count Cadastros (people) under the same Unidade filter, so that the card is not a duplicate of famílias.
35. As an Operador, I want Cestas básicas to count Núcleos with the cesta flag that have at least one Assistido in the selected Unidade, so that a house with a child in Centro appears when filtering Centro.
36. As an Operador, I want a chart of Assistidos por Unidade, so that I can see load per filial.
37. As an Operador, I want a chart of Alunos por Oficina (current enrolment), so that I can see workshop occupancy now.
38. As an Operador, I want a “Todas as Unidades” filter on painel and consulta, so that sede staff can see the whole organisation.
39. As an Operador, I want to download painel and consulta as PDF, as in the attached screens, so that I can share a snapshot without giving system access.
40. As an Operador, I want white page and card backgrounds, so that the UI is calmer than the beige prototype.
41. As an Operador, I want purple and beige only on details (labels, headers, accents, primary buttons), so that brand remains without painting the whole canvas.
42. As an Operador, I want primary actions (salvar, pesquisar, entrar) in purple rather than red, so that “download PDF” does not compete as the loudest object unless it is also restyled as a secondary beige/purple control.
43. As an Operador, I want 44px-class tap targets and 8px+ gaps on checkboxes and buttons, so that I can mark oficinas with a thumb.
44. As an Operador, I want visible focus rings and keyboard order that follows the form, so that I can register with a keyboard in the office.
45. As an Operador, I want save and search buttons to show a loading state and ignore double taps, so that I do not create two Cadastros by accident.
46. As an Operador, I want inline errors next to CPF when uniqueness fails, plus a focusable error summary on submit, so that I can jump to the problem.
47. As an Operador, I want oficina checkboxes labelled with the catalog from the attached screen (Jiu-jitsu, NPA 7 a 12 anos, Nação Esporte, Nação Cultura, Acessuas Trabalho, Juventude na Mesa), so that names stay stable in the painel.
48. As an Operador, I want Unidade options to come from a catalog (seeded), so that the select is not a free-text typo source.
49. As an Operador, I want an empty painel (zeros and empty charts) when there is no data, so that a new filial is not a broken page.
50. As an Operador, I want charts to not rely on colour alone (labels/values), so that beige/purple accents remain readable.
51. As a maintainer, I want a single Go HTTP API as the source of truth, so that React never invents KPI maths the API disagrees with.
52. As a maintainer, I want tests against that API (auth, cadastro, núcleo link, uniqueness, indicadores), so that agents can extend the system without a running browser.

## Implementation Decisions

- Greenfield repo `nacao-assistidos`: React PWA (SPA) + Go HTTP JSON API. Postgres as store. Online-first PWA (manifest + installable); no offline write queue in this PRD.
- Domain language is `CONTEXT.md` (Assistido, Aluno, Núcleo Familiar, Unidade, Oficina, Cesta Básica, Cadastro, Atendimento, Operador, Responsável Legal). Do not introduce Beneficiário as a person type.
- Authentication: Operador accounts provisioned (seed at least one); no public self-registration. Session via httpOnly cookie after login. All authenticated Operadores see all Unidades; Unidade filter is view-only, not ACL.
- Identity: Assistido has a server-generated id. CPF unique when non-empty; empty CPF allowed and never merged by name. Return visits: load Cadastro from consulta, then save.
- On first save from a blank form: create Núcleo (new, or attach to chosen existing), create Assistido, append Atendimento (date default today).
- On save of an existing Cadastro: update Assistido fields (nome, CPF, Unidade, Oficinas) and Núcleo fields (responsável, endereço, WhatsApp, cesta); always append a new Atendimento.
- Núcleo has no Unidade. Painel filter by Unidade: Assistidos with that Unidade; Núcleos/cestas included if at least one Assistido of the núcleo has that Unidade.
- Cesta Básica is one boolean on Núcleo (on the list now). Not two statuses (elegível vs recebe). Not counted from Atendimento text.
- Oficinas are the current set on Assistido, catalog fixed as in the attached form. Aluno = Assistido with ≥1 Oficina. Chart is current enrolment, not history.
- All resource fields optional at API and UI. Empty nome displays as “Sem nome”. No asterisks for required fields (none are required). Still validate CPF format/uniqueness only when CPF is provided.
- UI: white backgrounds for page and cards. Purple and beige only for labels, table headers, chart accents, chips, and buttons. Do not use the prototype’s full-page beige wash. SVG icons, not emoji. Visible labels; placeholders are hints only. Contrast ≥ 4.5:1 for body text on white.
- Screens: login; cadastro (create/edit); consulta; painel de indicadores — field set aligned with the three attached screens plus login and núcleo search.
- PDF export for painel and consulta as in the prototype (server-generated or client print of the same numbers the API returns; numbers must match the painel).
- Highest test seam: the Go HTTP API (JSON), including `GET` of indicadores. UI is a client of that seam, not a second source of KPI truth. Browser E2E is optional follow-up, not the primary contract.

## Testing Decisions

- Test external behaviour of the API: HTTP status, JSON body, auth cookie, uniqueness, aggregation. Do not test private function names or SQL.
- Good tests: create two Assistidos without CPF and assert two ids; link two Assistidos to one Núcleo and assert famílias=1, assistidos=2, cestas=1 when flag set; filter Unidade Centro when siblings split Centro/Norte; changing oficinas replaces enrolment in indicadores; second save of same id appends Atendimento and keeps one Assistido; duplicate CPF returns a field-level error; unauthenticated requests to cadastro/indicadores fail.
- Greenfield: no prior test suite. Establish API tests next to the Go service as the canonical suite.

## Out of Scope

- Self-service signup, OAuth, per-Unidade ACL, and Operador admin UI beyond a seed user.
- Automatic merge of Cadastros by name, phone, or address.
- Modelling Responsável as an Assistido or as a separate person entity.
- History of Oficina changes or cesta start/stop dates (only current flags/sets).
- Offline-first sync, native stores, or push notifications.
- Public website, donation portal, or WhatsApp bot.
- Treating “elegível” and “recebe” as two fields.
- Dark mode.
- Changing the oficina catalog without a later change.

## Further Notes

- Glossary: `CONTEXT.md` at repo root. Screenshots of the prototype (form, painel, consulta) are the visual field inventory; visual chrome is overridden (white canvas, purple/beige details).
- Proposed test seams (greenfield — no existing module seams): (1) Go JSON HTTP API as the contract; (2) indicadores as an API resource; (3) React PWA as consumer. If these seams are wrong, say so before implementation.
- Closed remaining grill recommendations that were not asked one-by-one: cesta is a single boolean on Núcleo; every ficha save appends Atendimento; auth is provisioned Operador + cookie session; PWA is installable but online-first; Unidade catalog is seeded, not free text.
