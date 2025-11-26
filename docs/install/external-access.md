# Access rmfakecloud from outside your local network safely

This guide explains how to get your reMarkable to access rmfakecloud from outside your local network, without exposing rmfakecloud directly to the internet or using VPNs, with a Cloudflare Tunnel and client authentication.

A domain name is needed for this setup, otherwise it's free. It will probably ask for a payment method to be added, but no charges will be made. The domain's nameservers also will need to point to Cloudflare.

!!! note
    Data is encrypted between the reMarkable tablet and Cloudflare and between Cloudflare and your local network, but **not on Cloudflare**.

## 1. Create a Cloudflare account and add a payment method

[Create a Cloudflare account](https://dash.cloudflare.com/sign-up) and add a [payment method](https://dash.cloudflare.com/?to=/:account/billing/payment-info). No charges will be made as everything is done within the free plan.

## 2. Add your domain to Cloudflare

In your [account home](https://dash.cloudflare.com/?to=/:account/home/domains), click "Onboard a domain" and enter your domain name. With "Quick scan for DNS records" enabled, Cloudflare will import your existing DNS records. Click "Continue" and select the free plan.

If you have an email set up with this domain or don't want other services to be proxied, set the corresponding records to "DNS only".

## 3. Change your domain nameservers

Follow the instructions you see on screen to change your domain's nameservers to the ones provided by Cloudflare. This is done on your domain registrar's website.

## 4. Set up a tunnel

In the Zero Trust dashboard (Cloudflare One), go to [**Networks > Connectors**](https://one.dash.cloudflare.com/?to=/:account/networks/connectors) and click "Create a tunnel".

Select "Cloudflared" tunnel type and follow the instructions to install a connector in your preferred environment. When you see an entry in the connectors list with status "Connected", that means the connector is running and ready.

In the "Route tunnel" step, set the subdomain you want to use to access rmfakecloud and your domain (for example, `rmfakecloud` and `example.com` to access the instance in `rmfakecloud.example.com`). For the rest of the guide, replace `rmfakecloud.example.com` with your chosen subdomain and domain.

In Service, set type to HTTPS and URL to a placeholder like `placeholder` (because we haven't set up authentication yet), then click "Save".

## 5. Generate a client certificate

In the domain dashboard, go to [**SSL/TLS > Client Certificates**](https://dash.cloudflare.com/?to=/:account/:zone/ssl-tls/client-certificates) and click "Create Certificate".

The default options are fine, so click "Create" (although 10 years validity is a lot, may be a good idea to set it to a shorter period).

Copy the text under "Certificate" and save it to a file `client.crt`, then copy the text under "Private Key" and save it to a file `client.key`. You can use whatever names you want for these files, just remember to use the names you choose instead of `client.crt` and `client.key`.

## 6. Get the certificate serial number

If you're on Linux or macOS (or if you have OpenSSL installed in Windows), you can run this command to get the get the serial number of the certificate:

```bash
openssl x509 -in client.crt -noout -serial | cut -d "=" -f2
```

On Windows, open properties of file `client.crt`, go to the "Details" tab, look for the "Serial number" field and copy its value.

The serial number has to be in **uppercase hexadecimal format without spaces or colons**, e.g. `123456789ABCDEF123456789ABCDEF123456789A`. If it's not, convert it before using it in the next step.

## 7. Set up a rule to block everything except requests with the client certificate

!!! note
    Cloudflare Zero Trust supports mTLS, but it's an [enterprise only feature](https://developers.cloudflare.com/learning-paths/mtls/concepts/mtls-cloudflare/). We can set a security rule to achieve the same effect.

Go to [**Security > Security rules**](https://dash.cloudflare.com/?to=/:account/:zone/security/security-rules), click "Create rule" and "Custom rules".

Set whatever name you want for the rule, e.g. "Require my rM client certificate".

Click "Edit expression" and enter the following expression, replacing `SERIAL_NUMBER` with the serial number obtained in the previous step and `rmfakecloud.example.com` with your chosen subdomain and domain:

```perl
(
    cf.tls_client_auth.cert_serial ne "SERIAL_NUMBER"
    or not cf.tls_client_auth.cert_verified
    or not cf.tls_client_auth.cert_presented
    or cf.tls_client_auth.cert_revoked
)
and http.host contains "rmfakecloud.example.com"
```

Info on these fields can be found [here](https://developers.cloudflare.com/ruleset-engine/rules-language/fields/reference/?field-category=mTLS).

Set the action to "Block" and click "Deploy".

## 8. Point the application route to your rmfakecloud instance

Back to the Zero Trust dashboard, go to [**Networks > Connectors**](https://one.dash.cloudflare.com/?to=/:account/networks/connectors), click on your tunnel and the "Edit" button.

Go to the "Published application routes" tab, click on the route you created earlier and edit it.

In the URL field, remove the placeholder set earlier (`placeholder`) and enter the local address of your rmfakecloud instance, e.g. `http://192.168.1.100:3000`. This will vary depending on your setup.

## 9. Check if it's working as expected

You can check if everything is working as expected running these commands in the directory where you saved the certificate and key files.

This command should return the normal rmfakecloud HTML response (won't work on Windows):

```bash
curl --cert client.crt --key client.key https://rmfakecloud.example.com
```

This command should return the HTML of a Cloudflare block page:

```bash
curl https://rmfakecloud.example.com
```

In Windows, if you have OpenSSL installed, you can to combine the certificate and key into a single `.pfx` file and then make the request with it. This should return the normal rmfakecloud HTML response:

```bash
openssl pkcs12 -export -out client.pfx -inkey client.key -in client.crt
curl --cert client.pfx https://rmfakecloud.example.com
```

## 10. Configure your reMarkable to use the new address and certificate

Copy the `client.crt` and `client.key` files to `/home/root/rmfakecloud` with your preferred method (`scp`, WinSCP, etc.).

Install `rmfakecloud-proxy` as the [installation instructions](../remarkable/setup.md#install-rmfakecloud-proxy) explain.

When asked for the cloud url, enter your public address. In this case, `https://rmfakecloud.example.com`.

In path to client certificate file, enter `client.crt`.

In path to client key file, enter `client.key`.

You should now be able to sync your reMarkable from outside your local network.
