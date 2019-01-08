### Introduction
This is an extension buildpack for .NET Core, HWC and binary buildpack. 
### Pre-requisites
CF CLI version 6.38 or later.
### How to use?
1. Deploy your app.
		For example:
cf push <app_name> -p <path to your app folder> -b dotnet_core_buildpack

2. Define the introscope service by running the following Cf CLI command. 
		
cf cups introscope -p {"url":"<value of the agent manager url>","agentManager.credential":"<credential only if connecting to SaaS EM>"}

Example for SaaS instance:
cf cups introscope -p '{"url":"https://665777.apm.cloud.ca.com:443","agentManager.credential":"0c3db026-a762-49e0-b467-52e5173de8db"}'

*On Prem example:
cf cups introscope -p {"url":”11.12.13.14:80”}’
Note: The service name must be introscope.

3. Bind your app to this service
		cf bind-service <app_name> introscope

4.Push the app again using the extension buildpack
cf push <app_name> -p <path to your app folder> -b https://github.com/CA-APM/ca_dotnet_core_ext_buildpack -b dotnet_core_buildpack
Note: You can keep the other arguments for the above command if you wanted to use for your application deployment.

5. Access your application 
5. Check the APM Server for agents to be connected and related performance metrics.

## Disclaimer
This buildpack is experimental and not yet intended for production use.
