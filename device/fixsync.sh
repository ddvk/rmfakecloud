# mark files as not synced
grep sync ~/.local/share/remarkable/xochitl/*.metadata -l | xargs sed -i 's/synced\": true/synced\": false/'
