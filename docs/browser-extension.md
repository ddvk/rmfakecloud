# Browser Extension support

As of commit [a62153e](https://github.com/ddvk/rmfakecloud/commit/a62153ef5c09145bbbf63079ccb7e77ed2024fbd),
the [Read on reMarkable](https://support.remarkable.com/hc/en-us/articles/360006830977-Read-on-reMarkable-Google-Chrome-Extension?preview%5Btheme_id%5D=210292689&preview_as_role=anonymous#)
Chrome extension is supported by rmfakecloud. Of course, the extension needs to be modified to connect to your cloud.

1. Install the extension from the Chrome Web Store.

2. Find the extension directory:
   - Go to `chrome://extensions` in your browser, find "Read on reMarkable" in the list, and take note of the ID.
     - Mine is `bfhkfdnddlhfippjbflipboognpdpoeh`.
   - Go to `chrome://version` in your browser, and take note of the "Profile Path".
     - Mine is `/home/murchu27/.config/BraveSoftware/Brave-Browser/Default` (Linux).
   - Your extension directory will be `{Profile Path}/Extensions/{ID}/{some version number}`.
     - Mine is `/home/murchu27/.config/BraveSoftware/Brave-Browser/Default/Extensions/bfhkfdnddlhfippjbflipboognpdpoeh/1.2.0_0`.

3. Within the extension directory, replace any instances of `https://internal.cloud.remarkable.com` and 
   `https://webapp-production-dot-remarkable-production.appspot.com/` with the URL of your cloud (same as the `STORAGE_URL` used by the tablet).
   You will need to do this in `manifest.json`, and any of the `.js` files.
   - If you are using Linux, you can save the below script as, e.g., `rmfakecloud-patch.sh`.
     ```
     mycloud=$1
     find ./ -type f -exec sed -i "s/https:\/\/internal.cloud.remarkable.com/$mycloud/g" {} \;
     find ./ -type f -exec sed -i "s/https:\/\/webapp-production-dot-remarkable-production.appspot.com/$mycloud/g" {} \;
     ```

   - Make this script executable with `chmod +x rmfakecloud-patch.sh`.
   - Run the script, passing in the URL of your cloud, e.g., `./rmfakecloud-patch.sh http://mycloud.com`.

4. Reload the extension in your browser:
   - Go to `chrome://extensions` in your browser. Enable "Developer mode" using the toggle at the top right.
   - Click the newly appeared "Load unpacked" button, a dialogue box will open.
   - Navigate to the extension directory (the same one as in the step above), and click "Open".

5. Try using the extension. Webpages should be sent to your tablet!

See [#67](https://github.com/ddvk/rmfakecloud/issues/67) for the original discussion of this feature.
