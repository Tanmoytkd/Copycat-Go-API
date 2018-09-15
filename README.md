# Copycat-Go-API
Code for the copycat API

#local:
curl -X POST -H "Content-Type: application/json" -d '{"Name": "your_name", "Email":"your_email", "Password":"your_password"}' localhost:8080/register
curl -X POST -H "Content-Type: application/json" -d '{"Email":"your_email", "Password":"your_password"}' localhost:8080/login
curl -X POST -H "Content-Type: application/json" -d '{"Token":"your_token_here"}' http://localhost:8080/logout


#remote:
curl -X POST -H "Content-Type: application/json" -d '{"Name": "Your Name", "Email":"your_email", "Password":"your_password"}' 40.117.123.41:8080/register
curl -X POST -H "Content-Type: application/json" -d '{"Email":"your_email", "Password":"your_password"}' 40.117.123.41:8080/login
curl -X POST -H "Content-Type: application/json" -d '{"Token":"your_token_here"}' http://40.117.123.41:8080/logout