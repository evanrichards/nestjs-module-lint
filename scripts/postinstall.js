#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { spawn } = require('child_process');

const packageJson = require('../package.json');
const version = packageJson.version;

function getPlatform() {
  const platform = process.platform;
  switch (platform) {
    case 'darwin':
      return 'darwin';
    case 'linux':
      return 'linux';
    case 'win32':
      return 'windows';
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }
}

function getArch() {
  const arch = process.arch;
  switch (arch) {
    case 'x64':
      return 'amd64';
    case 'arm64':
      return 'arm64';
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }
}

function getBinaryName() {
  const platform = getPlatform();
  const arch = getArch();
  const ext = platform === 'windows' ? '.exe' : '';
  return `nestjs-module-lint-${platform}-${arch}${ext}`;
}

function downloadBinary(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        https.get(response.headers.location, (redirectResponse) => {
          redirectResponse.pipe(file);
          file.on('finish', () => {
            file.close();
            resolve();
          });
        }).on('error', reject);
      } else if (response.statusCode === 200) {
        response.pipe(file);
        file.on('finish', () => {
          file.close();
          resolve();
        });
      } else {
        reject(new Error(`Failed to download: ${response.statusCode}`));
      }
    }).on('error', reject);
  });
}

async function buildFromSource() {
  console.log('Building from source...');
  const projectRoot = path.join(__dirname, '..');
  
  // Check if we have the source files
  const mainGoPath = path.join(projectRoot, 'main.go');
  if (!fs.existsSync(mainGoPath)) {
    throw new Error('Source files not available in npm package. Please download pre-built binaries.');
  }
  
  return new Promise((resolve, reject) => {
    const build = spawn('go', ['build', '-o', path.join(projectRoot, 'bin', 'nestjs-module-lint'), 'main.go'], {
      cwd: projectRoot,
      stdio: 'inherit'
    });

    build.on('close', (code) => {
      if (code !== 0) {
        reject(new Error(`Build failed with code ${code}`));
      } else {
        resolve();
      }
    });

    build.on('error', (err) => {
      reject(err);
    });
  });
}

async function install() {
  const binDir = path.join(__dirname, '..', 'bin');
  const binaryName = getBinaryName();
  const binaryPath = path.join(binDir, 'nestjs-module-lint');

  // Check if binary already exists (e.g., included in package)
  if (fs.existsSync(binaryPath)) {
    console.log('Binary already exists, skipping download/build.');
    // Make sure it's executable
    if (process.platform !== 'win32') {
      fs.chmodSync(binaryPath, 0o755);
    }
    return;
  }

  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  try {
    // First, try to download pre-built binary
    const downloadUrl = `https://github.com/evanrichards/nestjs-module-lint/releases/download/v${version}/${binaryName}`;
    console.log(`Downloading binary from ${downloadUrl}...`);
    
    await downloadBinary(downloadUrl, binaryPath);
    
    // Make binary executable
    if (process.platform !== 'win32') {
      fs.chmodSync(binaryPath, 0o755);
    }
    
    console.log('Binary downloaded successfully!');
  } catch (error) {
    console.warn(`Failed to download pre-built binary: ${error.message}`);
    console.log('Attempting to build from source...');
    
    try {
      await buildFromSource();
      console.log('Built from source successfully!');
    } catch (buildError) {
      console.error('Failed to build from source:', buildError.message);
      console.error('Please ensure Go is installed and try again.');
      process.exit(1);
    }
  }
}

// Only run install if this is being executed directly (not imported)
if (require.main === module) {
  install().catch((error) => {
    console.error('Installation failed:', error);
    process.exit(1);
  });
}