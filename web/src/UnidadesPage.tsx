import { atualizarUnidade, alternarAtivoUnidade, criarUnidade, listarUnidades } from "./api";
import { CatalogoAdminPage } from "./CatalogoAdminPage";
import { IconeUnidades } from "./Icones";

export function UnidadesPage() {
  return (
    <CatalogoAdminPage
      titulo="Unidades"
      icone={<IconeUnidades />}
      lede="Cadastre filiais e inative as que não devem mais aparecer no Cadastro."
      rotuloNome="Nome da Unidade"
      placeholder="ex.: Centro"
      tituloLista="Filiais cadastradas"
      vazio="Nenhuma Unidade ainda."
      rotuloCriar="Criar Unidade"
      listar={listarUnidades}
      criar={criarUnidade}
      atualizar={atualizarUnidade}
      alternarAtivo={alternarAtivoUnidade}
    />
  );
}
