About the project:
--
I created the server REST API application, which can be used for storing the data of any type and getting the files back. You can: delete your files by id or name, upload files, get the exact file or the list of your files.

Before the usage you need to be registered or logged in the system. After doing that you will be given the "session_id" which is needed for identification. So I use cookies in authentication and redis for storing it.
The user files stored in the server, they have some specific file structure. I use PostgreSQL for storing user's information and files as well. It should be noted your files called as messages.

System ensures security by watching for your credentials to make an access to your files. If you somehow get to know the directory name and file name of foreign user you still can't reach the access to this file.
I've made some test that shows functionality of the application.

Application's endpoints:
--

POST /register - to register in the system

POST /login - to login in the system if you already have an account


**Authorized access**:


POST /account/message - to upload your files

GET /account/message_list - watch the list of the files that you can get

DELETE /account/message_name - delete your file by name

DELETE /account/message_id -  delete your file by id

GET /account/messages/... - enter the name of the file that you want to get. You can get it in the specific structure by entering *"message_list"*.
