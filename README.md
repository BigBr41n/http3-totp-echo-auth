# TOTP authentication API using echo framework with HTTP3 

### Routes : 
- `/api/v1/auth/signup` 
- `/api/v1/auth/login`
- `/api/v1/auth/2fa/enable`
- `/api/v1/auth/verify-totp`


### flow : 

1. user signup 
2. user login 
    - if the user already enabled the 2fa , the return will be a temp jwt token 
    - if not he will get access token and refresh token back 

3. if the user didn't enable 2fa before then enable it using `/api/v1/auth/2fa/enable`
    - the result back will be secret & qr code 
    - store the TOTP in an authenticator app 

4. in the next login 
    - the user will get back a temp jwt 
    - the client send the temp jwt with the otp to `/api/v1/auth/verify-totp`
    - the response will be the access token , refresh token if the TOTP is valid 

