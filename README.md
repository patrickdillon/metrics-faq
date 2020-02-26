# metrics-faq
This repo provides context and background for the OpenShift Installer telemetry metrics. 

You can simply read through this document to understand how metrics aggregation works with the installer, or you can follow along if you want to play with a local setup.

## Prometheus
First you will need to have Prometheus running locally. You can [download it](https://prometheus.io/download/), as of today the latest release can be downloaded by running:

```
curl -LO https://github.com/prometheus/prometheus/releases/download/v2.16.0/prometheus-2.16.0.linux-amd64.tar.gz
```

Then unzip according to the [Getting Started guide](https://prometheus.io/docs/prometheus/latest/getting_started/):

```
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

## Standard Pushgateway: Basic Example

It might seem like a good idea for the installer to push metrics whose values have some significance beyond counting towards aggregation. For example, the sample's value could be the duration required for an install. To illustrate the need for the aggregation pushgateway, let's try running the [standard pushgateway](https://github.com/prometheus/pushgateway) as a counterexample. 

First, run the pushgateway on port 9091:
```
podman pull prom/pushgateway

podman run -d -p 9091:9091 prom/pushgateway
```

Then set Prometheus to listen on that port and run it:
```
sed -i 's/9090/9091/g' prometheus.yml
./prometheus --config.file=prometheus.yml
```

For the basic example, let's just use curl to push a metric called `cluster_installation_duration`. The following example could represent a cluster installation on a machine running Linux that took 30 minutes:
```
cat <<EOF | curl --data-binary @- http://localhost:9091/metrics/job/installation
# TYPE cluster_installation_duration gauge
cluster_installation_duration{os="linux"}30
EOF
```

Let's push some more samples. One representing an install that took 20 minutes, and another that took 35. Note that Prometheus scrapes the pushgateway every 15 seconds, so values could be overwritten if they arrive in between scrapes:
```
cat <<EOF | curl --data-binary @- http://localhost:9091/metrics/job/installation
# TYPE cluster_installation_duration gauge
cluster_installation_duration{os="linux"}20
EOF

# wait roughly 15 seconds

cat <<EOF | curl --data-binary @- http://localhost:9091/metrics/job/installation
# TYPE cluster_installation_duration gauge
cluster_installation_duration{os="linux"}35
EOF
```

When we look at the graph running on http://localhost:9090/graph we see a graph that looks something like this:
![Basic Example](assets/basic-pushgateway-screenshot.png)

Each change in the graph denotes a sample pushed to the gateway. But what does this graph tell us? Not much. An even more obvious example would be several consecutive samples with the same value (e.g. all of the installs fell into the 30 minute bucket). We would see a flat line and would have no indication that more than one sample was pushed. 

We must consider values that make sense as a function of time, hence the need for the aggregation gateway. 

## Aggregation Gateway: Basic Example

Instead of updating values in Prometheus, the aggregation pushgateway will add a sample to what is found in Prometheus. In this example, the values will always be 1 to represent a single installer invocation. Therefore a metric value would represent how many times the installer has been run fitting the criteria described by the labels.

Now stop the standard push gateway and stop Prometheus.

Start the aggregation pushgateway:
```
podman pull weaveworks/prom-aggregation-gateway

sudo podman run -d --rm -p 80:80 weaveworks/prom-aggregation-gateway
```

 We'll point Prometheus to scrape the aggregation pushgateway and restart it:

```
sed -i 's/9091/80/g' prometheus.yml
./prometheus --config.file=prometheus.yml

```


We'll push similar samples to last time, with a couple that fall into the same bucket:

```
echo 'cluster_installation_duration{os="linux", duration="30"}1' | curl --data-binary @- http://localhost/api/ui/metrics

# there is no need to wait between pushes 

echo 'cluster_installation_duration{os="linux", duration="30"}1' | curl --data-binary @- http://localhost/api/ui/metrics

echo 'cluster_installation_duration{os="linux", duration="30"}1' | curl --data-binary @- http://localhost/api/ui/metrics

echo 'cluster_installation_duration{os="linux", duration="25"}1' | curl --data-binary @- http://localhost/api/ui/metrics

```
On http://localhost:9090/graph (querying `cluster_installation_duration`) you should see something like this:
![Aggregation Basic Example](assets/aggregation-push-basic.png)

By using the aggregation pushgateway, we can see the growth over time and the current total of installations broken down by categories in labels.



