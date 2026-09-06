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
  const iconEle = document.querySelector(`head > link[rel='icon']`) as Element
  iconEle.setAttribute(`href`, `data:image/svg+xml,${faviconTemplate(emoji)}`)
}