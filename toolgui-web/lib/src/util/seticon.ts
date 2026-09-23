function faviconTemplate(icon: string) {
  return `
    <svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22>
      <text y=%22.9em%22 font-size=%2290%22>
        ${icon}
      </text>
    </svg>
  `.trim();
}

export function setIcon(emoji: string) {
  // Not every index.html ships a <link rel="icon">, so make one if missing.
  let iconEle = document.querySelector(`head > link[rel='icon']`)
  if (!iconEle) {
    iconEle = document.createElement('link')
    iconEle.setAttribute('rel', 'icon')
    document.head.appendChild(iconEle)
  }
  iconEle.setAttribute(`href`, `data:image/svg+xml,${faviconTemplate(emoji)}`)
}
