import assert from 'node:assert/strict';
import { test } from 'node:test';
import { readBuildRun, readSigningRun, validateManifest } from './provenance.mjs';

const sha = 'a'.repeat(40);
const fixture = () => ({
  id: 42, repository: { full_name: 'rongxinzy/engram' },
  head_repository: { full_name: 'rongxinzy/engram' },
  path: '.github/workflows/fork-release.yml', event: 'push',
  status: 'completed', conclusion: 'success', head_sha: sha,
  head_branch: 'v3.6.0-zhiyuan.1',
});
const apiFor = (run, status = 'ahead', tagSha = sha) => async (repo, suffix) => {
  assert.equal(repo, run.repository.full_name);
  if (suffix.startsWith('actions/runs/')) return run;
  if (suffix.startsWith('compare/')) return { status };
  if (suffix.startsWith('git/ref/tags/')) return { object: { type: 'commit', sha: tagSha } };
  throw new Error(`Unexpected provenance query: ${suffix}`);
};

test('build authority is the owned tagged workflow and immutable main ancestor', async () => {
  const build = await readBuildRun('engram', '42', apiFor(fixture()));
  assert.equal(build.sourceRepository, 'rongxinzy/engram');
  assert.equal(build.sourceSha, sha);
  for (const patch of [
    { repository: { full_name: 'z189yis/engram-cjk' } },
    { head_repository: { full_name: 'someone/engram' } },
    { path: '.github/workflows/release.yml' }, { event: 'pull_request' },
    { conclusion: 'failure' }, { status: 'in_progress' }, { id: 43 },
    { head_branch: 'main' }, { head_branch: 'v3.6.0' },
  ]) {
    const run = { ...fixture(), ...patch };
    await assert.rejects(readBuildRun('engram', '42', apiFor(run)));
  }
  await assert.rejects(readBuildRun('engram', '42', apiFor(fixture(), 'behind')));
  await assert.rejects(readBuildRun('engram', '42', apiFor(fixture(), 'ahead', 'b'.repeat(40))));
});

test('central signing authority must be the completed main workflow', async () => {
  const run = { ...fixture(), repository: { full_name: 'rongxinzy/RongxinAI' },
    head_repository: { full_name: 'rongxinzy/RongxinAI' },
    path: '.github/workflows/runtime-central-signing.yml',
    event: 'workflow_dispatch', head_branch: 'main' };
  assert.equal(await readSigningRun('42', apiFor(run)), run);
  await assert.rejects(readSigningRun('42', apiFor({ ...run, head_branch: 'release/test' })));
  await assert.rejects(readSigningRun('42', apiFor(run, 'diverged')));
});

test('signed manifest binds exact owner, source, signing run and both architectures', async () => {
  const build = await readBuildRun('engram', '42', apiFor(fixture()));
  const manifest = { schemaVersion: 1, signingRepository: 'rongxinzy/RongxinAI',
    signingRunId: '99', ...build,
    files: ['amd64', 'arm64'].map(arch => ({ name: `engram-windows-${arch}.exe`,
      unsignedSha256: 'c'.repeat(64), signedSha256: 'd'.repeat(64) })) };
  validateManifest(manifest, build, '99');
  for (const patch of [
    { sourceRepository: 'z189yis/engram-cjk' }, { sourceSha: 'b'.repeat(40) },
    { signingRunId: '98' }, { sourceTag: 'v3.6.0-zhiyuan.2' },
    { files: [manifest.files[0]] },
    { files: [manifest.files[0], manifest.files[0]] },
    { files: manifest.files.map(file => ({ ...file, signedSha256: 'invalid' })) },
  ]) assert.throws(() => validateManifest({ ...manifest, ...patch }, build, '99'));
});
