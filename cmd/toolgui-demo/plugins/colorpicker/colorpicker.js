// A plugin is a plain script the app serves. It runs in a sandboxed frame on
// an opaque origin, and talks to the app through window.toolgui.
(function () {
  var root = document.createElement('div')
  root.className = 'swatches'
  document.body.appendChild(root)

  // The plugin keeps no state of its own: what is selected comes from Go, so
  // a reconnect or a rerun shows the same thing.
  window.toolgui.onRender(function (props, theme) {
    document.body.className = theme === 'dark' ? 'dark' : ''

    root.innerHTML = ''

    var colors = props.colors || []
    for (var i = 0; i < colors.length; i++) {
      root.appendChild(swatch(colors[i], colors[i] === props.selected))
    }
  })

  function swatch(color, selected) {
    var button = document.createElement('button')
    button.className = selected ? 'swatch selected' : 'swatch'
    button.style.background = color
    button.title = color
    button.addEventListener('click', function () {
      window.toolgui.update({ color: color })
    })
    return button
  }

  // The frame gets none of the app's css, so it also has to say how tall it is.
  window.toolgui.autoHeight()
})()
