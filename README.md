This is a sample mutating admission controller for kubernetes that works on all kubernetes objects
It either adds the labels field if its missing from the metadata of the resource or mutates it based on the existence of a label called app on the resource
This controller ensures that no object is created without a defined label called app in the cluster so to enable identification and selection

To run the project clone the respository and build the image from the dockerfile given
Make a deployment and service in the cluster from the yaml files given
Then apply the configuration given in the repository to configure the webhook for the controller
