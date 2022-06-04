# Product Monitor Orchestrator

## Initial considerations

This project is part of a larger project that consists of a system that monitors products on popular e-commerce websites (such as amazon) and notifies an interested user through a Telegram bot. 
This specific part of the project contains the logic for crawling the products for registered users.
Currently, the other repositories containing the other parts of the application are still private.

## How it works

Pending products are fetched from a database (postgres) and each product is looked up concurrently.
All logs of the operation are currently being stored on a mongodb database.
Once the product data is retrieved from a website, if the current price is below the threshold specified by the user, a message is sent to a queue manager (rabbit mq). That message will be captured by a different application, which will alert the user through a Telegram bot that the product is available at the required price.