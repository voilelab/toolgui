import { AppConf } from "@toolgui-web/lib"

export async function getAppConf() {
    const resp = await fetch('/api/app')
    const appConf: AppConf = await resp.json()
    return appConf
}
