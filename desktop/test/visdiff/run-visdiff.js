import process from 'process'  // leave a little wiggle room for antialiasing inconsistencies
import fs from 'fs'
import path from 'path'
import {execSync} from 'child_process'
import gm from 'gm'

const BUCKET_S3 = 's3://keybase-app-visdiff'
const BUCKET_HTTP = 'http://keybase-app-visdiff.s3.amazonaws.com/'

function renderScreenshots (commitRange) {
  for (const commit of commitRange) {
    console.log(`Rendering screenshots of ${commit}`)
    execSync(`git checkout -f ${commit} && mkdir -p screenshots/${commit} && npm run render-screenshots -- screenshots/${commit}`)
  }
  execSync(`git checkout -f ${commitRange[1]}`)
}

function compareScreenshots (commitRange, diffDir, callback) {
  const results = {}

  execSync(`mkdir -p ${diffDir}`)

  const files = fs.readdirSync(`screenshots/${commitRange[0]}`)
  function compareNext () {
    const filename = files.pop()
    if (!filename || filename.startsWith('.')) {
      callback(results)
      return
    }

    const oldPath = `screenshots/${commitRange[0]}/${filename}`
    const newPath = `screenshots/${commitRange[1]}/${filename}`
    const diffPath = `${diffDir}/${filename}`
    const compareOptions = {
      tolerance: 1e-6,  // leave a little wiggle room for antialiasing inconsistencies
      file: diffPath
    }

    gm.compare(oldPath, newPath, compareOptions, (err, isEqual) => {
      if (err) {
        console.log(err)
        process.exit(1)
      }
      results[diffPath] = isEqual
      compareNext()
    })
  }
  compareNext()
}

if (process.argv.length !== 3) {
  console.log('Usage: node run-visdiff COMMIT1..COMMIT2')
  process.exit(1)
}

const commitRange = process.argv[2]
  .split(/\.{2,3}/)  // TRAVIS gives us ranges like START...END
  .map(s => s.substr(0, 12))  // trim the hashes a bit for shorter paths

// TODO fudging commit range for testing
commitRange[0] = 'b3904b9cb13c3fe81520c7632f0aa0cb826182fa'.substr(0, 12)

const diffDir = `screenshots/${Date.now()}-${commitRange[0]}-${commitRange[1]}`
renderScreenshots(commitRange)
compareScreenshots(commitRange, diffDir, results => {
  Object.keys(results).forEach(filePath => {
    if (results[filePath] === true) {
      fs.unlinkSync(filePath)
    } else {
      const filenameParts = path.parse(filePath, '.png')
      for (const commit of commitRange) {
        execSync(`cp screenshots/${commit}/${filenameParts.base} ${filenameParts.dir}/${filenameParts.name}-${commit}${filenameParts.ext}`)
      }
    }
  })

  const s3Env = {
    ...process.env,
    AWS_ACCESS_KEY_ID: process.env['VISDIFF_AWS_ACCESS_KEY_ID'],
    AWS_SECRET_ACCESS_KEY: process.env['VISDIFF_AWS_SECRET_ACCESS_KEY']
  }
  console.log(`Uploading screenshots to ${BUCKET_S3}`)
  execSync(`s3cmd put --acl-public -r ${diffDir} ${BUCKET_S3}`, {env: s3Env})
  console.log('Screenshots uploaded.')
})
