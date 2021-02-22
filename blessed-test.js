// * To "change the page", remove the screen's children, add new ones, render
// * Layout is like CSS' position: absolute. Calculate widths from content manually.
// * Setting the screen title is a nice touch.
// * Use screen.log and screen.debug since console.log doesn't go anywhere I can find.
// * TODO: Better text layout.
// * TODO: try out blessed-contrib grid or blessed's experimental Layout

// Mock data
var namespaces = [
  {name: 'default', age: '19d', status: 'active'},
  {name: 'kube-system', age: '19d', status: 'active'},
  {name: 'kube-public', age: '19d', status: 'active'}
]

function tableData(objs, props) {
  return objs.map((row) => {
    return props.map((prop) => row[prop])
  })
}

var blessed = require('blessed')
  , contrib = require('blessed-contrib')
  , screen = blessed.screen({ log: 'blessed.log' })

var table = contrib.table(
  { keys: true
  , mouse: true
  , vi: true
  , interactive: true
  //, search: function for /searchterms
  , label: 'Namespaces'
  , fg: 'white'
  , selectedFg: 'white'
  , selectedBg: 'blue'
  , top: 'center'
  , left: 'center'
  , width: '30%'
  , height: '30%'
  , border: {type: "line", fg: "green"}
  , columnSpacing: 20 //in chars
  , columnWidth: ['kube-system'.length, 'active'.length, 'age'.length] /*in chars*/
  })

//allow control the table with the keyboard
table.focus()

var properties = ['name', 'status', 'age']
table.setData(
{ headers: properties
, data: tableData(namespaces, properties) })

table.rows.on('select', (node, selectedIdx) => {
  // getText() is the whole row's text
  screen.log('select', node.getText(), selectedIdx)
  gotoDetail(selectedIdx)
})

// why append before setting data?
screen.append(table)

screen.key(['escape', 'q', 'C-c'], function(ch, key) {
  return process.exit(0);
});

screen.title = 'namespaces'
screen.log('foobar')
screen.render()

function gotoDetail(namespaceId) {

  // clear screen
  screen.children.forEach((n) => screen.remove(n))

  // Name:         hello
  // Labels:       <none>
  // Annotations:  <none>
  // Status:       Active
  //
  // No resource quota.
  //
  // No LimitRange resource.

  var ns = namespaces[namespaceId]
  screen.title = ns.name

  var box = blessed.box({
    top: 'center',
    left: 'center',
    width: '50%',
    height: '50%',
    content: 'Hello {bold}world{/bold}!\n' + `Name: ${ns.name}\nLabels: <none>\nAnnotations: <none>\nStatus: ${ns.status}\n\nNo resource quota.\n\nNo LimitRange.`,
    tags: true,
    border: {
      type: 'line'
    },
    style: {
      fg: 'white',
      bg: 'magenta',
      border: {
        fg: '#f0f0f0'
      },
      hover: {
        bg: 'green'
      }
    }
  });

  screen.append(box)
  screen.render()

}
