#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

// Path to the binary
const binaryPath = path.join(__dirname, 'bin', 'nestjs-module-lint');

// Spawn the binary with all arguments passed through
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  shell: false
});

// Handle exit
child.on('exit', (code) => {
  process.exit(code);
});

// Handle errors
child.on('error', (err) => {
  if (err.code === 'ENOENT') {
    console.error('Error: nestjs-module-lint binary not found.');
    console.error('Please run "npm install" to download the binary.');
  } else {
    console.error('Error executing nestjs-module-lint:', err.message);
  }
  process.exit(1);
});