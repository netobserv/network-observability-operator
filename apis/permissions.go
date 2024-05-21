package api

// Operator permissions
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces;services;serviceaccounts;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;clusterroles;rolebindings;roles,verbs=get;list;create;delete;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;list;update;watch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors;flowmetrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/status;flowmetrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flows.netobserv.io,resources=flowcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=hostnetwork,verbs=use
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=list;create;update;watch
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=list;get;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors;prometheusrules,verbs=get;create;delete;update;patch;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions;networks,verbs=get;list;watch
//+kubebuilder:rbac:groups=loki.grafana.com,resources=network,resourceNames=logs,verbs=get;create
//+kubebuilder:rbac:urls="/metrics",verbs=get

// FLP specific permissions (the operator needs to have at least the same permissions in order to grant them)
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=core,resources=pods;nodes,verbs=get;list;watch

// Console plugin permissions
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
