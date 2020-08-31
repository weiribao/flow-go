package chunks

import (
	"fmt"
	"sort"

	"github.com/dapperlabs/flow-go/crypto/hash"
	"github.com/dapperlabs/flow-go/crypto/random"
	"github.com/dapperlabs/flow-go/engine/verification/utils"
	chunkmodels "github.com/dapperlabs/flow-go/model/chunks"
	"github.com/dapperlabs/flow-go/model/encoding"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module/mempool"
	"github.com/dapperlabs/flow-go/module/mempool/stdmap"
)

// DefaultChunkAssignmentAlpha is the default number of verifiers that should be
// assigned to each chunk.
// DISCLAIMER: alpha down there is not a production-level value
const DefaultChunkAssignmentAlpha = 1

// PublicAssignment implements an instance of the Public Chunk Assignment
// algorithm for assigning chunks to verifier nodes in a deterministic but
// unpredictable manner. It implements the ChunkAssigner interface.
type PublicAssignment struct {
	alpha       int // used to indicate the number of verifiers that should be assigned to each chunk
	assignments mempool.Assignments
}

// NewPublicAssignment generates and returns an instance of the Public Chunk
// Assignment algorithm. Parameter alpha is the number of verifiers that should
// be assigned to each chunk.
func NewPublicAssignment(alpha int) (*PublicAssignment, error) {
	// TODO to have limit of assignment mempool as a parameter (2703)
	assignment, err := stdmap.NewAssignments(1000)
	if err != nil {
		return nil, fmt.Errorf("could not create an assignment mempool: %w", err)
	}
	return &PublicAssignment{
		alpha:       alpha,
		assignments: assignment,
	}, nil
}

// Size returns number of assignments
func (p *PublicAssignment) Size() uint {
	return p.assignments.Size()
}

// Assign generates the assignment using the execution result to seed the RNG.
func (p *PublicAssignment) Assign(verifiers flow.IdentityList, result *flow.ExecutionResult) (*chunkmodels.Assignment, error) {
	rng, err := utils.NewChunkAssignmentRNG(result)
	if err != nil {
		return nil, fmt.Errorf("could not generate random generator: %w", err)
	}

	// computes a finger print for identities||chunks
	ids := verifiers.NodeIDs()
	hash, err := fingerPrint(ids, result.Chunks, rng, p.alpha)
	if err != nil {
		return nil, fmt.Errorf("could not compute hash of identifiers: %w", err)
	}

	// checks cache against this assignment
	assignmentFingerprint := flow.HashToID(hash)
	a, exists := p.assignments.ByID(assignmentFingerprint)
	if exists {
		return a, nil
	}

	// otherwise, it computes the assignment and caches it for future calls
	a, err = chunkAssignment(ids, result.Chunks, rng, p.alpha)
	if err != nil {
		return nil, fmt.Errorf("could not complete chunk assignment: %w", err)
	}

	// adds assignment to mempool
	added := p.assignments.Add(assignmentFingerprint, a)
	if !added {
		return nil, fmt.Errorf("could not add generated assignment to mempool")
	}

	return a, nil
}

// GetAssignedChunks returns the list of result chunks assigned to a specific
// verifier.
func (p *PublicAssignment) GetAssignedChunks(verifierID flow.Identifier, assignment *chunkmodels.Assignment, result *flow.ExecutionResult) (flow.ChunkList, error) {
	// indices of chunks assigned to verifier
	chunkIndices := assignment.ByNodeID(verifierID)

	// chunks keeps the list of chunks assigned to the verifier
	chunks := make(flow.ChunkList, 0, len(chunkIndices))
	for _, index := range chunkIndices {
		chunk, ok := result.Chunks.ByIndex(index)
		if !ok {
			return nil, fmt.Errorf("chunk out of range requested: %v", index)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// chunkAssignment implements the business logic of the Public Chunk Assignment algorithm and returns an
// assignment object for the chunks where each chunk is assigned to alpha-many verifier node from ids list
func chunkAssignment(ids flow.IdentifierList, chunks flow.ChunkList, rng random.Rand, alpha int) (*chunkmodels.Assignment, error) {
	if len(ids) < alpha {
		return nil, fmt.Errorf("not enough verification nodes for chunk assignment: %d, minumum should be %d", len(ids), alpha)
	}

	// creates an assignment
	assignment := chunkmodels.NewAssignment()

	// permutes the entire slice
	err := rng.Shuffle(len(ids), ids.Swap)
	if err != nil {
		return nil, fmt.Errorf("shuffling verifiers failed: %w", err)
	}
	t := ids

	for i := 0; i < chunks.Len(); i++ {
		assignees := make([]flow.Identifier, 0, alpha)
		if len(t) >= alpha { // More verifiers than required for this chunk
			assignees = append(assignees, t[:alpha]...)
			t = t[alpha:]
		} else { // Less verifiers than required for this chunk
			assignees = append(assignees, t...) // take all remaining elements from t

			// now, we need `still` elements from a new shuffling round:
			still := alpha - len(assignees)
			t = ids[:ids.Len()-len(assignees)] // but we exclude the elements we already picked from the population
			err := rng.Samples(len(t), still, t.Swap)
			if err != nil {
				return nil, fmt.Errorf("sampling verifiers failed: %w", err)
			}

			// by adding `still` elements from new shuffling round: we have alpha assignees for the current chunk
			assignees = append(assignees, t[:still]...)

			// we have already assigned the first `still` elements in `ids`
			// note that remaining elements ids[still:] still need shuffling
			t = ids[still:]
			err = rng.Shuffle(len(t), t.Swap)
			if err != nil {
				return nil, fmt.Errorf("shuffling verifiers failed: %w", err)
			}
		}
		// extracts chunk by index
		chunk, ok := chunks.ByIndex(uint64(i))
		if !ok {
			return nil, fmt.Errorf("chunk out of range requested: %v", i)
		}
		assignment.Add(chunk, assignees)
	}
	return assignment, nil
}

// Fingerprint computes the SHA3-256 hash value of the inputs to the assignment algorithm:
// - sorted version of identifier list
// - sorted version of chunk list
// - internal state of random generator
// - alpha
// the generated fingerprint is deterministic in the set of aforementioned parameters
func fingerPrint(ids flow.IdentifierList, chunks flow.ChunkList, rng random.Rand, alpha int) (hash.Hash, error) {
	// sorts and encodes ids
	sort.Sort(ids)
	encIDs, err := encoding.DefaultEncoder.Encode(ids)
	if err != nil {
		return nil, fmt.Errorf("could not encode identifier list: %w", err)
	}

	// sorts and encodes chunks
	sort.Sort(chunks)
	encChunks, err := encoding.DefaultEncoder.Encode(chunks)
	if err != nil {
		return nil, fmt.Errorf("could not encode chunk list: %w", err)
	}

	// encodes random generator
	encRng := rng.State()

	// encodes alpha parameteer
	encAlpha, err := encoding.DefaultEncoder.Encode(alpha)
	if err != nil {
		return nil, fmt.Errorf("could not encode alpha: %w", err)
	}

	// computes and returns hash(encIDs || encChunks || encRng || encAlpha)
	hasher := hash.NewSHA3_256()
	_, err = hasher.Write(encIDs)
	if err != nil {
		return nil, fmt.Errorf("could not hash ids: %w", err)
	}
	_, err = hasher.Write(encChunks)
	if err != nil {
		return nil, fmt.Errorf("could not hash chunks: %w", err)
	}

	_, err = hasher.Write(encRng)
	if err != nil {
		return nil, fmt.Errorf("could not random generator: %w", err)
	}

	_, err = hasher.Write(encAlpha)
	if err != nil {
		return nil, fmt.Errorf("could not hash alpha: %w", err)
	}

	return hasher.SumHash(), nil
}
