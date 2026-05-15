import fs from 'fs';
import path from 'path';

const root = path.resolve('src/bank');
const badOpen = '<' + 'motion';
const badClose = '</' + 'motion>';

function walk(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      walk(full);
    } else if (entry.name.endsWith('.jsx')) {
      let content = fs.readFileSync(full, 'utf8');
      content = content.split(badClose).join('</div>');
      content = content.split(badOpen).join('<div');
      fs.writeFileSync(full, content);
    }
  }
}

walk(root);
console.log('Fixed bank JSX tags');
