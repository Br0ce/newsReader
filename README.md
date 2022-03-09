# NewsReader

Side project to test some patterns around event driven design/ event sourcing.

--- 

News is all around us, feeding our opinions. Because news is constantly changing, it is hard to keep track of the most
important issues, how they are represented, and what patterns emerge as they spread across the news landscape.

The main goal of NewsReader is to provide means to monitor and analyze news content over time.

NewsReader periodically crawls various news sources for content and extracts features by applying pretrained models
from `Huggingface`. The enriched news content is added to `OpenSearch` to enable full text search and visualizations.

## News Content

The examined unit of NewsReader is an article. News articles evolve over time. Its content is expanded or revised and
its category may also change. For example, if its topic no longer meets the breaking news criterion. For this app, the
article entity is defined by its host and title.

## Components

* `Collector`: A Collector crawls articles from all provided Crawlers and publishes the results to a queue.
* `Operator`: An Operator consumes articles from a queue, applies all provided Processors and republishes them.
* [pytorch/serve](https://github.com/pytorch/serve)
* [openSearch](https://github.com/opensearch-project/OpenSearch)
* [EventstoreDB](https://github.com/EventStore/EventStore)

## Design

Collector and Operator communicate by consuming or publishing articles. Every article has its own queue and propagates
through the system by changing its state in an event-sourcing manner.

Crawlers used by the Collector and Processors used by the Operator are used concurrently whenever possible. In addition,
most processors delegate the actual computation to `pytorch/serve` synchronously over http.
