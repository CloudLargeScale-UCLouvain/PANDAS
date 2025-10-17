package peersim.kademlia.das.operations;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import peersim.kademlia.das.Block;
import peersim.kademlia.das.KademliaCommonConfigDas;
import peersim.kademlia.das.MissingNode;
import peersim.kademlia.das.Sample;
import peersim.kademlia.das.SearchTable;
import peersim.kademlia.operations.FindOperation;

public abstract class SamplingOperation extends FindOperation {

  protected SearchTable searchTable;
  protected int samplesCount = 0;
  protected boolean completed;
  protected boolean isValidator;
  protected MissingNode callback;
  protected Block currentBlock;
  protected int aggressiveness;
  // protected HashSet<BigInteger> queried;
  protected int strategy;
  protected int timeout;
  protected int timesIncreased;
  protected BigInteger radiusValidator, radiusNonValidator;

  protected LinkedHashMap<BigInteger, Node> nodes;
  protected HashMap<BigInteger, FetchingSample> samples;

  protected List<BigInteger> askNodes;

  public SamplingOperation(
      BigInteger srcNode,
      BigInteger destNode,
      long timestamp,
      Block block,
      boolean isValidator,
      int numValidators) {
    super(srcNode, destNode, timestamp);
    completed = false;
    this.isValidator = isValidator;
    currentBlock = block;
    strategy = KademliaCommonConfigDas.validatorStrategy;
    if (strategy == 3) {
      timeout = peersim.kademlia.Timeout.TIMEOUT;
    }
    radiusNonValidator =
        currentBlock.computeRegionRadius(KademliaCommonConfigDas.NUM_SAMPLE_COPIES_PER_PEER);
    samples = new HashMap<>();
    nodes = new LinkedHashMap<>();
    this.available_requests = 0;
    aggressiveness = 0;
    askNodes = new ArrayList<>();
    timesIncreased = 0;
  }

  public SamplingOperation(
      BigInteger srcNode,
      BigInteger destNode,
      long timestamp,
      Block block,
      boolean isValidator,
      int numValidators,
      MissingNode callback) {
    super(srcNode, destNode, timestamp);
    samples = new HashMap<>();
    nodes = new LinkedHashMap<>();
    completed = false;
    this.isValidator = isValidator;
    this.callback = callback;
    currentBlock = block;
    this.available_requests = 0;
    aggressiveness = 0;
    radiusValidator =
        currentBlock.computeRegionRadius(
            KademliaCommonConfigDas.NUM_SAMPLE_COPIES_PER_PEER, numValidators);
    radiusNonValidator =
        currentBlock.computeRegionRadius(KademliaCommonConfigDas.NUM_SAMPLE_COPIES_PER_PEER);
    askNodes = new ArrayList<>();
    timesIncreased = 0;
  }

  public abstract boolean completed();

  public void updateTimeout(int timeout) {
    this.timeout = timeout;
  }

  public int getTimeout() {
    return this.timeout;
  }

  public int getStrategy() {
    return this.strategy;
  }

  public BigInteger[] getSamples() {
    List<BigInteger> result = new ArrayList<>();

    for (FetchingSample sample : samples.values()) {
      if (!sample.isDownloaded()) result.add(sample.getId());
    }

    return result.toArray(new BigInteger[0]);
  }

  public BigInteger getRadiusValidator() {
    return radiusValidator;
  }

  public BigInteger getRadiusNonValidator() {
    return radiusNonValidator;
  }

  protected abstract void createNodes();

  // === Paper Algorithm ===
  public BigInteger[] doSampling() {

    aggressiveness += KademliaCommonConfigDas.row_column_sampling_aggressiveness_step;
    for (Node n : nodes.values()) n.setAgressiveness(aggressiveness);
    List<BigInteger> result = new ArrayList<>();
    for (Node n : nodes.values()) {
      if (!n.isBeingAsked() && n.getScore() > 0) { // break;
        n.setBeingAsked(true);
        this.available_requests++;
        for (FetchingSample s : n.getSamples()) {
          s.addFetchingNode(n);
        }
        result.add(n.getId());
      }
    }

    return result.toArray(new BigInteger[0]);
  }

  public BigInteger[] doRowColumnSampling() {
    aggressiveness += KademliaCommonConfigDas.row_column_sampling_aggressiveness_step;
    for (Node n : nodes.values()) n.setAgressiveness(aggressiveness);
    List<BigInteger> result = new ArrayList<>();
    for (Node n : nodes.values()) {
      if (!n.isBeingAsked() && n.getScore() > 0) { // break;
        n.setBeingAsked(true);
        this.available_requests++;
        for (FetchingSample s : n.getSamples()) {
          s.addFetchingNode(n);
        }
        result.add(n.getId());
      }
    }

    return result.toArray(new BigInteger[0]);
  }

  public BigInteger[] doRandomSampling() {

    aggressiveness += KademliaCommonConfigDas.random_sampling_aggressiveness_step;
    if (KademliaCommonConfigDas.validatorStrategy == 1) {
      aggressiveness = 0;
    }
    for (Node n : nodes.values()) n.setAgressiveness(aggressiveness);
    List<BigInteger> result = new ArrayList<>();
    for (Node n : nodes.values()) {
      if (!n.isBeingAsked() && n.getScore() > 0) { // break;
        n.setBeingAsked(true);
        this.available_requests++;
        for (FetchingSample s : n.getSamples()) {
          s.addFetchingNode(n);
        }
        result.add(n.getId());
      }
    }

    return result.toArray(new BigInteger[0]);
  }

  public abstract void elaborateResponse(Sample[] sam, BigInteger node);

  public abstract void elaborateResponse(Sample[] sam);

  public int samplesCount() {
    return samplesCount;
  }
}
