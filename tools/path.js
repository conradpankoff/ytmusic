const fs = require('fs')

function find(subject, value, path, onMatch) {
  if (subject === value) {
    onMatch(path);
  } else if (typeof subject === 'string' && typeof value === 'string' && subject.indexOf(value) !== -1) {
    onMatch(path);
  }

  if (typeof subject !== 'object') {
    return
  }

  if (Array.isArray(subject)) {
    for (let i = 0; i < subject.length; i++) {
      find(subject[i], value, path.concat([i]), onMatch)
    }
  } else {
    for (const k in subject) {
      find(subject[k], value, path.concat([k]), onMatch)
    }
  }
}

const data = JSON.parse(fs.readFileSync('playlist.json'))

const paths = []

find(data, "UCUl32_nMbQndCK39ra63E9Q", [], (path) => void paths.push(path.join('.')))

console.log(paths)
