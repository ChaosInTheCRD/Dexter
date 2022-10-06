# Dexter

<center> Don't expect your engineers to know why or how they should declare their image references for better security, leave it to Dexter. </center>

<p align="center">
  <img src="./logo/dexter.png" height="256" width="256" alt="dexter project logo" />
</p>

⚠️  **WARNING: Dexter is not yet ready for prime time. There is still work to do and it might yet be buggy, so please refrain from trusting it to be 100% reliable. If you think it is awesome and want it in your production CI, hop in and submit some PRs! It is a small tool so it hopefully won't be too long until Dexter is ready for action!** ⚠️

Dexter can be placed into your CI processes to scrape through the repository and find files that contain image references, with the aim of pinning them to the immutable digest.

## What can Dexter Find?

Dexter currently has a small selection of parsers that are able to find image references within the file types they are designed to assess. These are:
- Kubernetes (currently supports Pod, Deployment, Replicaset, Statefulset, Daemonset, Job, CronJob)
- Dockerfile

## What Problem does this solve?

Let's say I have a file with an image declared like so:

```yaml
image: 'grafana/grafana:4.2.0'
```

The `4.2.0` is what is called an image tag. Image tags (as you might have guessed), make it quick and easy to point to an image, but it is not an immutable identity to an image itself. The tag points to a long (not very nice to look at) sha256 hash known as an image digest:

```
7ff7f9b2501a5d55b55ce3f58d21771b1c5af1f2a4ab7dbf11bef7142aae7033
```

Image digests are great because they are the immutable identity that an image tag points to. But why do I care? A-ha! This is where Dexter comes in.

the `4.2.0` tag points to the `7ff7f9b2501a5d55b55ce3f58d21771b1c5af1f2a4ab7dbf11bef7142aae7033` digest identity of the image sat in the registry. However, nothing is stopping me as someone with write access to the registry, to point the `4.2.0` tag at a different image. In most cases this could be innocent changes to the image, but to do such a thing is *never* good practice. In the worst case, this could be a malicious modification to the image so it contains a cypto-miner, or in other words; a supply-chain attack.

Once you have applied Dexter against your repository, your image reference (taking the grafana image shown above as an example) will look like:

```yaml
image: index.docker.io/grafana/grafana@sha256:7ff7f9b2501a5d55b55ce3f58d21771b1c5af1f2a4ab7dbf11bef7142aae7033
```

Now if this image is running in your estate, you will not be susceptible to an image tag getting hijacked and you will no longer get hit by the dreaded crypto-miner under this use case! woohoo!

## Quickstart

To quickly test out the tool on your repository, simply execute the following command from within the Dexter repository:

```bash
go run manipulate --directory <your-repository-directory>
```

## Dexter Configuration

Dexter has a simple configuration file for allowing you to specify how you would like it to behave when it executes against your repository. *Please note that this is currently very basic, and there is an intention to make this more fully featured in the near future*.

### Parsers

This allows you to select the parsers that you wish to use:

```yaml
parsers:
- "dockerfile"
- "kubernetes"
```

Of course there are not many parsers yet 👀 but hopefully this will become handy in the near future.

### Ignore (DO NOT USE THIS YET IT IS NOT READY)

The intention here is that you might want to tell Dexter of specific files, and even specific references within those files that you do not want it to touch. Being the enthusiastic nerd that Dexter is, he'll ensure that he leaves these references alone to the letter.
