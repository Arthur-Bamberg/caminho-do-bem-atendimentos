import { atualizarOficina, alternarAtivoOficina, criarOficina, listarOficinas } from "./api";
import { CatalogoAdminPage } from "./CatalogoAdminPage";
import { IconeOficinas } from "./Icones";

export function OficinasPage() {
  return (
    <CatalogoAdminPage
      titulo="Oficinas"
      icone={<IconeOficinas />}
      lede="Cadastre atividades vigentes e inative as que não devem mais aparecer no Cadastro."
      rotuloNome="Nome da Oficina"
      placeholder="ex.: Jiu-jitsu"
      tituloLista="Atividades cadastradas"
      vazio="Nenhuma Oficina ainda."
      rotuloCriar="Criar Oficina"
      listar={listarOficinas}
      criar={criarOficina}
      atualizar={atualizarOficina}
      alternarAtivo={alternarAtivoOficina}
    />
  );
}
