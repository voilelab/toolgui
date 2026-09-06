export interface AppConf {
  page_names: string[]
  page_confs: { [page_name: string]: any }

  main_container_id: string,
  sidebar_container_id: string,

  hash_page_name_mode: boolean,
}