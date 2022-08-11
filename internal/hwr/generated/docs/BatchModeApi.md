# {{classname}}

All URIs are relative to *https://cloud.myscript.com*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Batch**](BatchModeApi.md#Batch) | **Post** /api/v4.0/iink/batch | Recognize iink

# **Batch**
> Batch(ctx, body, applicationKey, hmac, accept, optional)
Recognize iink

This endpoint sends an json object representing an iink input to the recognition engine and sends back a payload whose format depends on the requested mime type. In case of error, an error object in json format is returned. That is why it is mandatory to ask in the Accept header the requested format + application/json

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**BatchInput**](BatchInput.md)|  | 
  **applicationKey** | **string**| The application key that was given during registration. | 
  **hmac** | **string**| The HMAC signature of the payload, using the secret hmac key that was given during registration and SHA-512 algorithm. See https://developer.myscript.com/support/account/registering-myscript-cloud/#computing-the-hmac-value | 
  **accept** | [**[]string**](string.md)| A comma-separated list of mime types for the response. Must contain application/json as the error response is a json object. The first suitable mime type for the recognition type will be used as content type for the response | 
 **optional** | ***BatchModeApiBatchOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a BatchModeApiBatchOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **userId** | **optional.**| The userId (only for specific customers) | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

