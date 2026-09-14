import assert from "node:assert/strict";
import test from "node:test";

import { isFirefox, requestLocalCandidateAccess } from "./firefoxUsb.js";

test("recognizes Firefox without treating Chromium as Firefox", () => {
  assert.equal(isFirefox("Mozilla/5.0 Firefox/142.0"), true);
  assert.equal(isFirefox("Mozilla/5.0 Chrome/142.0 Safari/537.36"), false);
});

test("requests audio permission and immediately stops every track", async () => {
  const stopped = [];
  const mediaDevices = {
    async getUserMedia(constraints) {
      assert.deepEqual(constraints, { audio: true });
      return {
        getTracks: () => [
          { stop: () => stopped.push("first") },
          { stop: () => stopped.push("second") },
        ],
      };
    },
  };

  assert.equal(await requestLocalCandidateAccess(mediaDevices), true);
  assert.deepEqual(stopped, ["first", "second"]);
});

test("fails clearly when media permissions are unavailable", async () => {
  await assert.rejects(
    requestLocalCandidateAccess(undefined),
    /Media permissions are unavailable/,
  );
});
