export function isFirefox(userAgent) {
  return /Firefox\//.test(userAgent);
}

export async function requestLocalCandidateAccess(mediaDevices) {
  if (!mediaDevices?.getUserMedia) {
    throw new Error("Media permissions are unavailable in this browser context");
  }

  const stream = await mediaDevices.getUserMedia({ audio: true });
  try {
    return true;
  } finally {
    for (const track of stream.getTracks()) track.stop();
  }
}
