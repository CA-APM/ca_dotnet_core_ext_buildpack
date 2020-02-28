## :warning: This is no longer being maintained. The buildpack is avaiable in PCF tile.

# CA APM .NET Buildpack For CloudFoundry

# Description
This is a CloudFoundry buildpack for .NET Core, HWC and binary buildpack.

## Short Description

This is a CloudFoundry buildpack for .NET Core, HWC and binary buildpack.

## APM version
CA APM 10.7 and DXI

## Supported third party versions
CF CLI version 6.38 or later.

## Limitations
n/a

## License
[Eclipse Public License - v 1.0](LICENSE)

# Installation Instructions

## Prerequisites
CF CLI version 6.38 or later.

## Dependencies
CA APM 10.7 and DXI

## Installation
1. Deploy your application. For example:

 `cf push <app_name> -p <path to your app folder> -b dotnet_core_buildpack`

2. Define the Introscope service by running the following CF CLI command:

 `cf cups introscope -p {"url":"<value of the agent manager url>","agentManager.credential":"<credential only if connecting to SaaS EM>"}`

 a. Example for CA APM SaaS (DXI):

 `cf cups introscope -p '{"url":"https://665777.apm.cloud.ca.com:443","agentManager.credential":"0c3db026-a762-49e0-b467-52e5173de8db"}'`

 b. Example for CA APM on premise:

   `cf cups introscope -p {"url":”11.12.13.14:80”}’`

 *Note: The service name must be "introscope".*

3. Bind your app to this service:

 `cf bind-service <app_name> introscope`

4. Push the app again using the extension buildpack

 `cf push <app_name> -p <path to your app folder> -b https://github.com/CA-APM/ca_dotnet_core_ext_buildpack -b dotnet_core_buildpack`

 *Note: You can keep the other arguments for the above command if you wanted to use for your application deployment.*

5. Access your application
6. Check the APM Server for agents to be connected and related performance metrics.

## Debugging and Troubleshooting
This buildpack is experimental and not yet intended for production use. Let us know about any issues or questions you encounter by opening a [GitHub issue](https://github.com/CA-APM/ca_dotnet_core_ext_buildpack/issues).

## Support
This document and associated tools are made available from CA Technologies as examples and provided at no charge as a courtesy to the CA APM Community at large. This resource may require modification for use in your environment. However, please note that this resource is not supported by CA Technologies, and inclusion in this site should not be construed to be an endorsement or recommendation by CA Technologies. These utilities are not covered by the CA Technologies software license agreement and there is no explicit or implied warranty from CA Technologies. They can be used and distributed freely amongst the CA APM Community, but not sold. As such, they are unsupported software, provided as is without warranty of any kind, express or implied, including but not limited to warranties of merchantability and fitness for a particular purpose. CA Technologies does not warrant that this resource will meet your requirements or that the operation of the resource will be uninterrupted or error free or that any defects will be corrected. The use of this resource implies that you understand and agree to the terms listed herein.

Although these utilities are unsupported, please let us know if you have any problems or questions by adding a comment to the CA APM Community Site area where the resource is located, so that the Author(s) may attempt to address the issue or question.

Unless explicitly stated otherwise this extension is only supported on the same platforms as the APM core agent. See [APM Compatibility Guide](http://www.ca.com/us/support/ca-support-online/product-content/status/compatibility-matrix/application-performance-management-compatibility-guide.aspx).

### Support URL
https://github.com/CA-APM/ca_dotnet_core_ext_buildpack/issues

# Contributing
The [CA APM Community](https://communities.ca.com/community/ca-apm) is the primary means of interfacing with other users and with the CA APM product team.  The [developer subcommunity](https://communities.ca.com/community/ca-apm/ca-developer-apm) is where you can learn more about building APM-based assets, find code examples, and ask questions of other developers and the CA APM product team.

If you wish to contribute to this or any other project, please refer to [easy instructions](https://communities.ca.com/docs/DOC-231150910) available on the CA APM Developer Community.

## Categories
Cloud


# Change log
Changes for each version of the extension.

Version | Author | Comment
--------|--------|--------
1.0 | CA Technologies | First version of the extension.
