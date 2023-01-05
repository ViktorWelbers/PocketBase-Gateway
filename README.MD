# PocketBase Gateway

This is an Authentication-Gateway that uses project PocketBase for authentication and user management.

It was used on Azure for the authentication of all other services in my image generation deep learning
project.

# Functionality

Besides the usual authentication via PocketBase, it also provides 4 other endpoints, one beeing the
`/health` endpoint, which is used by the load balancer to check if the service is up and running.

The other 3 endpoints are the following:

1. `api/send_message` - This endpoint can be used to send a prompt to an azure queue which is then used to trigger an
   machine learning model which can create a different/adjusted version of an image. The endpoint will return a simple
   JSON object with the message id of the message that was sent to the queue, which is used to identify the transformed
   version of the image.
   The parameters it expects are the following.

   ```json
   {
        "prompt" : "The prompt for the image generation",
        "prompt_guidance" : "The prompt guidance for the image generation",
        "strength" : "The strength of the prompt",
        "uuid" : "The uuid of the image that is used to identify the image in the database",
        "model_type" : "The model type that should be used for the image generation"
   }
   ```
2. `api/get_image` - This endpoint can be used the try to get an image that has been edited by the machine learning
   model. Since this is an async messenger pattern (due to the long-lrunning nature of the machine learning model when
   the system is under load), the image might not be ready yet and the endpoint will return a
   simple string with the value `"Message not ready yet"`. The frontend I designed used this endpoint to check if the
   image is ready and if it is, it will display the image. The parameters it expects are the following:
    ```json
    {
          "message_id" : "The message id that is returned by the send_message endpoint, which is used to identify the image"
    }
    ```

3. `api/check_auth` - This endpoint can be used to check if the user is authenticated. It will return a simple JSON
   object with the value `{"client_id": <USERNAME>}` if the user is authenticated via a token that PocketBase has
   created
   and a Statuscode `403` with the message `Only auth records can access this endpoint` if the user is not
   authenticated.
